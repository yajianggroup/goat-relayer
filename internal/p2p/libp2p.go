package p2p

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	tcp "github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	log "github.com/sirupsen/logrus"

	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/state"
)

const (
	handshakeProtocol  = "/goat/voter/handshake/1.0.0"
	messageProtocol    = "/goat/voter/message/1.0.0"
	expectedHandshake  = "goatvotergoatbest"
	messageTopicName   = "gossip-topic"
	heartbeatTopicName = "heartbeat-topic"
	privKeyFile        = "node_private_key.pem"
)

var messageTopic *pubsub.Topic
var hbTopic *pubsub.Topic
var pubsubService *pubsub.PubSub
var currentMessageSub *pubsub.Subscription
var currentHBSub *pubsub.Subscription

type LibP2PService struct {
	state *state.State
	// Add context cancellation for goroutines
	topicsCancel context.CancelFunc
	topicsCtx    context.Context
}

func NewLibP2PService(state *state.State) *LibP2PService {
	return &LibP2PService{
		state: state,
	}
}

func (lp *LibP2PService) Start(ctx context.Context) {
	node, ps, err := createNodeWithPubSub(ctx)
	if err != nil {
		log.Fatalf("Failed to create libp2p node: %v", err)
	}

	// Store global references for topic reconnection
	pubsubService = ps

	printNodeAddrInfo(node)

	node.SetStreamHandler(protocol.ID(handshakeProtocol), func(s network.Stream) {
		log.Println("New handshake stream")
		handleHandshake(s, node)
		s.Close()
	})

	go lp.connectToBootNodes(ctx, node)

	// Initialize topics and subscriptions
	err = lp.initializeTopics(ctx, node)
	if err != nil {
		log.Fatalf("Failed to initialize topics: %v", err)
	}

	// Enhanced network diagnosis with topic recovery
	go lp.startNetworkDiagnosisWithRecovery(ctx, node)

	go func() {
		time.Sleep(5 * time.Second)
		msg := Message[any]{
			RequestId:   "1",
			MessageType: MessageTypeUnknown,
			Data:        "Hello, goat voter libp2p PubSub network with handshake!",
		}

		// Log detailed network state before publishing
		peers := node.Network().Peers()
		log.Infof("🔗 Attempting to publish test message with %d connected peers", len(peers))
		for _, peerID := range peers {
			conn := node.Network().ConnsToPeer(peerID)
			if len(conn) > 0 {
				log.Infof("  📡 Connected to peer: %s via %s", peerID, conn[0].RemoteMultiaddr())
			}
		}

		if err := PublishMessage(ctx, msg); err != nil {
			log.Errorf("❌ Failed to publish test message: %v", err)
		} else {
			log.Infof("✅ Test message published successfully to %d peers", len(peers))
		}
	}()

	<-ctx.Done()

	log.Info("LibP2PService is stopping...")
	lp.Stop()
	if err := node.Close(); err != nil {
		log.Errorf("Error closing libp2p node: %v", err)
	}
	log.Info("LibP2PService has stopped.")
}

// initializeTopics sets up topic subscriptions and message handlers
func (lp *LibP2PService) initializeTopics(ctx context.Context, node host.Host) error {
	var err error

	// Create a cancellable context for this topics initialization
	lp.topicsCtx, lp.topicsCancel = context.WithCancel(ctx)

	// Join message topic
	messageTopic, err = pubsubService.Join(messageTopicName)
	if err != nil {
		return fmt.Errorf("failed to join message topic: %v", err)
	}

	// Subscribe to message topic
	currentMessageSub, err = messageTopic.Subscribe()
	if err != nil {
		return fmt.Errorf("failed to subscribe to message topic: %v", err)
	}

	// Join heartbeat topic
	hbTopic, err = pubsubService.Join(heartbeatTopicName)
	if err != nil {
		return fmt.Errorf("failed to join heartbeat topic: %v", err)
	}

	// Subscribe to heartbeat topic
	currentHBSub, err = hbTopic.Subscribe()
	if err != nil {
		return fmt.Errorf("failed to subscribe to heartbeat topic: %v", err)
	}

	// Start message handlers with the topics-specific context
	go lp.handlePubSubMessages(lp.topicsCtx, currentMessageSub, node)
	go lp.handleHeartbeatMessages(lp.topicsCtx, currentHBSub, node)
	go startHeartbeat(lp.topicsCtx, node)

	log.Infof("✅ Topics initialized successfully")
	return nil
}

// reconnectTopics handles topic reconnection when subscriptions are lost
func (lp *LibP2PService) reconnectTopics(ctx context.Context, node host.Host) error {
	log.Warnf("🔄 Attempting to reconnect topics...")

	// Cancel existing goroutines first
	if lp.topicsCancel != nil {
		log.Debugf("Cancelling existing topic goroutines...")
		lp.topicsCancel()
		// Wait a moment for goroutines to exit
		time.Sleep(1 * time.Second)
	}

	// Close existing subscriptions if they exist
	if currentMessageSub != nil {
		currentMessageSub.Cancel()
		currentMessageSub = nil
	}
	if currentHBSub != nil {
		currentHBSub.Cancel()
		currentHBSub = nil
	}

	// Close existing topics to allow re-joining
	if messageTopic != nil {
		messageTopic.Close()
		messageTopic = nil
	}
	if hbTopic != nil {
		hbTopic.Close()
		hbTopic = nil
	}

	// Wait a moment for cleanup
	time.Sleep(2 * time.Second)

	// Reinitialize topics
	return lp.initializeTopics(ctx, node)
}

// ForceTopicReconnection allows manual triggering of topic reconnection
// This is useful for external monitoring or manual intervention
func (lp *LibP2PService) ForceTopicReconnection(ctx context.Context, node host.Host) error {
	log.Warnf("🔄 Manual topic reconnection triggered")
	return lp.reconnectTopics(ctx, node)
}

func (lp *LibP2PService) connectToBootNodes(ctx context.Context, node host.Host) {
	bootNodeAddrs := strings.Split(config.AppConfig.Libp2pBootNodes, ",")
	var wg sync.WaitGroup

	for _, addr := range bootNodeAddrs {
		if addr == "" {
			continue
		}
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					log.Infof("Context cancelled, stopping connection attempts to %s", addr)
					return
				default:
					err := lp.connectToBootNode(ctx, node, addr)
					if err != nil {
						log.Errorf("Failed to connect to bootnode %s: %v", addr, err)
						time.Sleep(10 * time.Second)
					} else {
						log.Infof("Successfully connected to bootnode %s", addr)
						lp.monitorConnection(ctx, node, addr)
						return
					}
				}
			}
		}(addr)
	}

	wg.Wait()

	peers := node.Network().Peers()
	if len(peers) == 0 {
		log.Warnf("No peers, please check your network connection")
	} else {
		log.Infof("Connected to %d peers", len(peers))
	}
}

func (lp *LibP2PService) connectToBootNode(ctx context.Context, node host.Host, addr string) error {
	// Generate multiple transport addresses for better connectivity
	peerAddrs, err := generateTransportAddresses(addr)
	if err != nil {
		return fmt.Errorf("failed to generate transport addresses for %s: %v", addr, err)
	}

	var lastErr error
	var connected bool
	var connectedPeerID peer.ID

	for i, peerAddr := range peerAddrs {
		log.Debugf("Attempting connection to %s via %s", peerAddr.ID, peerAddr.Addrs[0])

		if err := node.Connect(ctx, *peerAddr); err != nil {
			lastErr = err
			log.Debugf("Failed to connect via %s: %v", peerAddr.Addrs[0], err)
			if i < len(peerAddrs)-1 {
				continue // Try next transport
			}
		} else {
			log.Infof("Successfully connected to %s via %s", peerAddr.ID, peerAddr.Addrs[0])
			connected = true
			connectedPeerID = peerAddr.ID
			break // Connection successful
		}
	}

	// If all transports failed, return the last error
	if !connected {
		return fmt.Errorf("failed to connect to bootnode %s (tried %d transports): %v", addr, len(peerAddrs), lastErr)
	}

	// Perform handshake with the successfully connected peer
	s, err := node.NewStream(ctx, connectedPeerID, protocol.ID(handshakeProtocol))
	if err != nil {
		return fmt.Errorf("failed to create handshake stream with %s: %v", addr, err)
	}
	defer s.Close()

	err = sendHandshake(s)
	if err != nil {
		return fmt.Errorf("failed to send handshake to %s: %v", addr, err)
	}

	return nil
}

func (lp *LibP2PService) monitorConnection(ctx context.Context, node host.Host, addr string) {
	// Generate transport addresses for comprehensive monitoring
	peerAddrs, err := generateTransportAddresses(addr)
	if err != nil {
		log.Errorf("Failed to generate transport addresses for %s: %v", addr, err)
		return
	}

	if len(peerAddrs) == 0 {
		log.Errorf("No valid addresses generated for monitoring %s", addr)
		return
	}

	// Use the first address for peer ID (all have the same peer ID)
	targetPeerID := peerAddrs[0].ID
	var consecutiveFailures int
	const maxConsecutiveFailures = 5

	log.Infof("Starting connection monitoring for peer %s", targetPeerID)

	for {
		select {
		case <-ctx.Done():
			log.Infof("Context cancelled, stopping monitoring of %s", addr)
			return
		default:
			// Enhanced connection status check
			isConnected := lp.isReallyConnected(node, targetPeerID)

			if !isConnected {
				consecutiveFailures++
				log.Warnf("Disconnected from %s (failure #%d), attempting to reconnect", addr, consecutiveFailures)

				// Clean up any stale connections before reconnecting
				lp.cleanupStaleConnections(node, targetPeerID)

				// Attempt reconnection with exponential backoff
				backoffDelay := time.Duration(consecutiveFailures) * 5 * time.Second
				if backoffDelay > 30*time.Second {
					backoffDelay = 30 * time.Second
				}

				err := lp.connectToBootNode(ctx, node, addr)
				if err != nil {
					log.Errorf("Failed to reconnect to %s: %v", addr, err)
					if consecutiveFailures >= maxConsecutiveFailures {
						log.Warnf("Max consecutive failures reached for %s, using longer backoff", addr)
						time.Sleep(backoffDelay)
					} else {
						time.Sleep(backoffDelay)
					}
					continue
				}

				log.Infof("Successfully reconnected to %s", addr)
				consecutiveFailures = 0 // Reset failure counter on success
			} else {
				consecutiveFailures = 0 // Reset failure counter when connected
			}

			// Check connection status every 20 seconds
			time.Sleep(20 * time.Second)
		}
	}
}

// isReallyConnected performs a more thorough connectivity check
func (lp *LibP2PService) isReallyConnected(node host.Host, peerID peer.ID) bool {
	// Check basic connectivity
	if node.Network().Connectedness(peerID) != network.Connected {
		return false
	}

	// Check if we have active connections
	conns := node.Network().ConnsToPeer(peerID)
	if len(conns) == 0 {
		return false
	}

	// Verify at least one connection is actually open
	for _, conn := range conns {
		if conn.IsClosed() {
			continue
		}
		// Connection exists and is open
		return true
	}

	return false
}

// cleanupStaleConnections removes any stale or broken connections
func (lp *LibP2PService) cleanupStaleConnections(node host.Host, peerID peer.ID) {
	conns := node.Network().ConnsToPeer(peerID)
	for _, conn := range conns {
		if conn.IsClosed() {
			log.Debugf("Cleaning up closed connection to %s", peerID)
			// Connection is already closed, libp2p should handle cleanup
			continue
		}
	}

	// Force disconnect to ensure clean state
	if node.Network().Connectedness(peerID) != network.NotConnected {
		log.Debugf("Force disconnecting from %s for clean reconnection", peerID)
		_ = node.Network().ClosePeer(peerID)
	}
}

func parseAddr(addrStr string) (*peer.AddrInfo, error) {
	addr, err := multiaddr.NewMultiaddr(addrStr)
	if err != nil {
		return nil, err
	}
	return peer.AddrInfoFromP2pAddr(addr)
}

// generateTransportAddresses creates multiple transport addresses from a single address
func generateTransportAddresses(originalAddr string) ([]*peer.AddrInfo, error) {
	// Parse the original address to extract IP and peer ID
	addr, err := multiaddr.NewMultiaddr(originalAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid multiaddr: %v", err)
	}

	// Extract IP address
	var ip string
	var ipVersion string
	if ipValue, err := addr.ValueForProtocol(multiaddr.P_IP4); err == nil {
		ip = ipValue
		ipVersion = "ip4"
	} else if ipValue, err := addr.ValueForProtocol(multiaddr.P_IP6); err == nil {
		ip = ipValue
		ipVersion = "ip6"
	} else {
		return nil, fmt.Errorf("no IP address found in multiaddr")
	}

	// Extract port (try TCP first, then UDP)
	var port string
	if tcpPort, err := addr.ValueForProtocol(multiaddr.P_TCP); err == nil {
		port = tcpPort
	} else if udpPort, err := addr.ValueForProtocol(multiaddr.P_UDP); err == nil {
		port = udpPort
	} else {
		return nil, fmt.Errorf("no port found in multiaddr")
	}

	// Extract peer ID
	peerIDStr, err := addr.ValueForProtocol(multiaddr.P_P2P)
	if err != nil {
		return nil, fmt.Errorf("no peer ID found in multiaddr: %v", err)
	}

	addresses := []string{
		fmt.Sprintf("/%s/%s/tcp/%s/p2p/%s", ipVersion, ip, port, peerIDStr),
	}

	var peerAddrs []*peer.AddrInfo
	for _, addrStr := range addresses {
		peerInfo, err := parseAddr(addrStr)
		if err != nil {
			log.Debugf("Failed to parse address %s: %v", addrStr, err)
			continue
		}
		peerAddrs = append(peerAddrs, peerInfo)
	}

	if len(peerAddrs) == 0 {
		return nil, fmt.Errorf("no valid transport addresses generated from %s", originalAddr)
	}

	return peerAddrs, nil
}

func sendHandshake(s network.Stream) error {
	handshakeMsg := []byte(expectedHandshake)

	_, err := s.Write(handshakeMsg)
	if err != nil {
		return err
	}

	// verify handshake
	buf := make([]byte, len(handshakeMsg))
	_, err = s.Read(buf)
	if err != nil {
		return err
	}

	if !bytes.Equal(buf, handshakeMsg) {
		return fmt.Errorf("invalid handshake response")
	}

	return nil
}

func createNodeWithPubSub(ctx context.Context) (host.Host, *pubsub.PubSub, error) {
	privKey, err := loadOrCreatePrivateKey(privKeyFile)
	if err != nil {
		return nil, nil, err
	}

	listenAddr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", config.AppConfig.Libp2pPort)

	addrsOpt := libp2p.AddrsFactory(func(in []multiaddr.Multiaddr) (out []multiaddr.Multiaddr) {
		for _, a := range in {
			if manet.IsPublicAddr(a) || manet.IsPrivateAddr(a) {
				// exclude 0.0.0.0 / 127.0.0.1 / link-local
				if !manet.IsIPLoopback(a) && !manet.IsIP6LinkLocal(a) && !manet.IsIPUnspecified(a) {
					out = append(out, a)
				}
			}
		}
		// "/ip4/52.88.11.99/tcp/4001,/ip4/54.91.64.7/tcp/4001"
		for _, s := range strings.FieldsFunc(os.Getenv("EXTERNAL_P2P_ADDR"), func(r rune) bool { return r == ',' || r == ' ' }) {
			if m, err := multiaddr.NewMultiaddr(strings.TrimSpace(s)); err == nil {
				out = append(out, m)
			} else {
				log.Warnf("bad multiaddr %q: %v", s, err)
			}
		}
		return
	})

	// Enhanced libp2p options for better NAT traversal and connectivity
	opts := []libp2p.Option{
		libp2p.Identity(privKey),
		// Enable multiple transports for better connectivity
		libp2p.Transport(tcp.NewTCPTransport),
		// Listen on TCP
		libp2p.ListenAddrStrings(listenAddr),
		addrsOpt,
		// Security and muxer options
		libp2p.DefaultSecurity,
		libp2p.DefaultMuxers,
		// Connection manager for better peer management
		libp2p.DefaultConnectionManager,
		// Peerstore configuration
		libp2p.DefaultPeerstore,
	}

	node, err := libp2p.New(opts...)
	if err != nil {
		return nil, nil, err
	}

	ps, err := pubsub.NewGossipSub(ctx, node, pubsub.WithPeerExchange(true))
	if err != nil {
		return nil, nil, err
	}

	return node, ps, nil
}

func loadOrCreatePrivateKey(fileName string) (crypto.PrivKey, error) {
	dbDir := config.AppConfig.DbDir
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	pemPath := filepath.Join(dbDir, fileName)
	if _, err := os.Stat(pemPath); err == nil {
		privKeyBytes, err := ioutil.ReadFile(pemPath)
		if err != nil {
			return nil, err
		}
		privKey, err := crypto.UnmarshalPrivateKey(privKeyBytes)
		if err != nil {
			return nil, err
		}
		return privKey, nil
	}

	privKey, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 2048, rand.Reader)
	if err != nil {
		return nil, err
	}

	privKeyBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(pemPath, privKeyBytes, 0600); err != nil {
		return nil, err
	}

	return privKey, nil
}

func printNodeAddrInfo(node host.Host) {
	addrs := node.Addrs()
	peerID := node.ID().String()

	for _, addr := range addrs {
		fullAddr := fmt.Sprintf("%s/p2p/%s", addr, peerID)
		log.Infof("Listening on: %s", fullAddr)
	}
}

// startNetworkDiagnosisWithRecovery monitors network status and automatically recovers from topic subscription loss
func (lp *LibP2PService) startNetworkDiagnosisWithRecovery(ctx context.Context, node host.Host) {
	ticker := time.NewTicker(60 * time.Second) // Every minute
	defer ticker.Stop()

	var lastTopicRecovery time.Time
	const topicRecoveryInterval = 2 * time.Minute // Minimum interval between topic recovery attempts

	for {
		select {
		case <-ticker.C:
			peers := node.Network().Peers()
			connectedPeers := 0

			var messageTopicPeers, hbTopicPeers []peer.ID
			if messageTopic != nil {
				messageTopicPeers = messageTopic.ListPeers()
			} else {
				log.Warnf("⚠️ Message topic is nil, cannot publish message")
			}
			if hbTopic != nil {
				hbTopicPeers = hbTopic.ListPeers()
			} else {
				log.Warnf("⚠️ Heartbeat topic is nil, cannot publish heartbeat")
			}

			log.Infof("📊 Network Status Report:")
			log.Infof("  🔗 Total connected peers: %d", len(peers))
			log.Infof("  📢 Peers in message topic: %d", len(messageTopicPeers))
			log.Infof("  💓 Peers in heartbeat topic: %d", len(hbTopicPeers))

			for _, peerID := range peers {
				conns := node.Network().ConnsToPeer(peerID)
				if len(conns) > 0 {
					connectedPeers++
					conn := conns[0]
					log.Infof("  📡 Peer %s: %s (direction: %s)",
						peerID.ShortString(),
						conn.RemoteMultiaddr(),
						conn.Stat().Direction)
				}
			}

			log.Infof("  📈 Active connections: %d", connectedPeers)

			// Check for potential issues and auto-recovery
			if len(peers) == 0 {
				log.Warnf("⚠️  No connected peers! Check your network configuration")
			} else if len(messageTopicPeers) == 0 || len(hbTopicPeers) == 0 {
				log.Errorf("🚨 CRITICAL: Topic subscriptions lost! Message topic peers: %d, Heartbeat topic peers: %d",
					len(messageTopicPeers), len(hbTopicPeers))

				// Auto-recovery: attempt to reconnect topics if enough time has passed
				if time.Since(lastTopicRecovery) > topicRecoveryInterval {
					log.Warnf("🔄 Initiating automatic topic recovery...")
					if err := lp.reconnectTopics(ctx, node); err != nil {
						log.Errorf("❌ Topic recovery failed: %v", err)
					} else {
						log.Infof("✅ Topic recovery completed successfully")
						lastTopicRecovery = time.Now()
					}
				} else {
					timeUntilNext := topicRecoveryInterval - time.Since(lastTopicRecovery)
					log.Warnf("⏰ Topic recovery on cooldown, next attempt in %v", timeUntilNext.Round(time.Second))
				}
			} else if len(messageTopicPeers) < len(peers) || len(hbTopicPeers) < len(peers) {
				log.Warnf("⚠️  Some peers not in topics - Message: (%d/%d), Heartbeat: (%d/%d)",
					len(messageTopicPeers), len(peers), len(hbTopicPeers), len(peers))
			} else {
				log.Infof("✅ All network connections and topic subscriptions are healthy")
			}

		case <-ctx.Done():
			return
		}
	}
}

// Stop gracefully stops the LibP2P service and cancels all goroutines
func (lp *LibP2PService) Stop() {
	log.Infof("🛑 Stopping LibP2P service...")

	// Cancel topic goroutines
	if lp.topicsCancel != nil {
		log.Debugf("Cancelling topic goroutines...")
		lp.topicsCancel()
	}

	// Close subscriptions
	if currentMessageSub != nil {
		currentMessageSub.Cancel()
		currentMessageSub = nil
	}
	if currentHBSub != nil {
		currentHBSub.Cancel()
		currentHBSub = nil
	}

	// Close topics
	if messageTopic != nil {
		messageTopic.Close()
		messageTopic = nil
	}
	if hbTopic != nil {
		hbTopic.Close()
		hbTopic = nil
	}

	log.Infof("✅ LibP2P service stopped")
}
