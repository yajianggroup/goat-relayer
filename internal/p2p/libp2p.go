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
	"github.com/libp2p/go-libp2p/p2p/host/autonat"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	tcp "github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
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

type LibP2PService struct {
	state *state.State
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
	printNodeAddrInfo(node)

	node.SetStreamHandler(protocol.ID(handshakeProtocol), func(s network.Stream) {
		log.Println("New handshake stream")
		handleHandshake(s, node)
		s.Close()
	})

	go lp.connectToBootNodes(ctx, node)

	messageTopic, err = ps.Join(messageTopicName)
	if err != nil {
		log.Fatalf("Failed to join message topic: %v", err)
	}

	sub, err := messageTopic.Subscribe()
	if err != nil {
		log.Fatalf("Failed to subscribe to message topic: %v", err)
	}

	hbTopic, err := ps.Join(heartbeatTopicName)
	if err != nil {
		log.Fatalf("Failed to join heartbeat topic: %v", err)
	}

	hbSub, err := hbTopic.Subscribe()
	if err != nil {
		log.Fatalf("Failed to subscribe to heartbeat topic: %v", err)
	}

	go lp.handlePubSubMessages(ctx, sub, node)
	go lp.handleHeartbeatMessages(ctx, hbSub, node)
	go startHeartbeat(ctx, node, hbTopic)

	// Enhanced network diagnosis
	go lp.startNetworkDiagnosis(ctx, node, messageTopic)

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
	if err := node.Close(); err != nil {
		log.Errorf("Error closing libp2p node: %v", err)
	}
	log.Info("LibP2PService has stopped.")
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

	// Try each transport protocol in order of preference: QUIC first (better NAT traversal), then TCP
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
	peerInfo, err := parseAddr(addr)
	if err != nil {
		log.Errorf("Failed to parse bootnode address %s: %v", addr, err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Infof("Context cancelled, stopping monitoring of %s", addr)
			return
		default:
			if node.Network().Connectedness(peerInfo.ID) != network.Connected {
				log.Warnf("Disconnected from %s, attempting to reconnect", addr)
				err := lp.connectToBootNode(ctx, node, addr)
				if err != nil {
					log.Errorf("Failed to reconnect to %s: %v", addr, err)
					time.Sleep(5 * time.Second)
					continue
				}
				log.Infof("Successfully reconnected to %s", addr)
			}
			time.Sleep(20 * time.Second)
		}
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
// Priority order: QUIC first (better NAT traversal), then TCP
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

	// Generate addresses for different transports (QUIC first for better NAT traversal)
	addresses := []string{
		fmt.Sprintf("/%s/%s/udp/%s/quic-v1/p2p/%s", ipVersion, ip, port, peerIDStr),
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
	listenAddrQUIC := fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", config.AppConfig.Libp2pPort)

	// Enhanced libp2p options for better NAT traversal and connectivity
	opts := []libp2p.Option{
		libp2p.Identity(privKey),
		// Enable multiple transports for better connectivity
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(quic.NewTransport),
		// Listen on both TCP and QUIC
		libp2p.ListenAddrStrings(listenAddr, listenAddrQUIC),
		// Enable AutoNAT for automatic public address discovery
		libp2p.EnableAutoNATv2(),
		// Enable NAT port mapping (UPnP, NAT-PMP)
		libp2p.EnableNATService(),
		// Enable circuit relay v2 (both as client and limited relay)
		libp2p.EnableRelayService(),
		// Enable autorelay for automatic relay usage when behind NAT
		libp2p.EnableAutoRelayWithStaticRelays([]peer.AddrInfo{}),
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

	// Start AutoNAT service for NAT detection
	if _, err := autonat.New(node); err != nil {
		log.Warnf("Failed to create AutoNAT client: %v", err)
	}

	ps, err := pubsub.NewGossipSub(ctx, node)
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
		log.Infof("Bootnode address: %s", fullAddr)
	}
}

func (lp *LibP2PService) startNetworkDiagnosis(ctx context.Context, node host.Host, topic *pubsub.Topic) {
	ticker := time.NewTicker(60 * time.Second) // Every minute
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			peers := node.Network().Peers()
			connectedPeers := 0
			topics := topic.ListPeers()

			log.Infof("📊 Network Status Report:")
			log.Infof("  🔗 Total connected peers: %d", len(peers))
			log.Infof("  📢 Peers in message topic: %d", len(topics))

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

			// Warn about potential issues
			if len(peers) == 0 {
				log.Warnf("⚠️  No connected peers! Check your network configuration")
			} else if len(topics) < len(peers) {
				log.Warnf("⚠️  Not all peers are in message topic (%d/%d)", len(topics), len(peers))
			}

		case <-ctx.Done():
			return
		}
	}
}
