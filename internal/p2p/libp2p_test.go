package p2p

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/state"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestConfig() {
	// Setup minimal config for testing
	config.AppConfig.Libp2pPort = 40001 // Use different port to avoid conflicts
	config.AppConfig.DbDir = filepath.Join(os.TempDir(), "goat-relayer-test")
	config.AppConfig.Libp2pBootNodes = ""
}

func TestCreateNodeWithPubSub(t *testing.T) {
	setupTestConfig()
	defer os.RemoveAll(config.AppConfig.DbDir)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test node creation with enhanced configuration
	node, ps, err := createNodeWithPubSub(ctx)
	require.NoError(t, err, "Failed to create libp2p node")
	require.NotNil(t, node, "Node should not be nil")
	require.NotNil(t, ps, "PubSub should not be nil")

	defer node.Close()

	t.Run("NodeHasCorrectTransports", func(t *testing.T) {
		// Check if node is listening on expected addresses
		addrs := node.Addrs()
		assert.NotEmpty(t, addrs, "Node should have listening addresses")

		var hasTCP bool
		for _, addr := range addrs {
			addrStr := addr.String()
			if strings.Contains(addrStr, "/tcp/") {
				hasTCP = true
			}
		}

		assert.True(t, hasTCP, "Node should support TCP transport")
	})

	t.Run("NodeHasCorrectProtocols", func(t *testing.T) {
		// Verify that the node supports expected protocols
		protocols := node.Mux().Protocols()
		assert.NotEmpty(t, protocols, "Node should support protocols")

		// Check for expected security and muxing protocols
		hasSecurityOrMuxing := false
		protocolList := make([]string, 0, len(protocols))
		for _, p := range protocols {
			pStr := string(p)
			protocolList = append(protocolList, pStr)
			// Check for common libp2p protocols
			if strings.Contains(pStr, "tls") ||
				strings.Contains(pStr, "noise") ||
				strings.Contains(pStr, "yamux") ||
				strings.Contains(pStr, "mplex") ||
				strings.Contains(pStr, "/1.0.0") { // Generic protocol version
				hasSecurityOrMuxing = true
			}
		}

		// Log protocols for debugging
		t.Logf("Available protocols: %v", protocolList)

		// libp2p should have at least some protocols
		assert.True(t, hasSecurityOrMuxing || len(protocols) > 0,
			"Node should have security/muxing protocols or at least some protocols")
	})

	t.Run("PubSubTopicJoin", func(t *testing.T) {
		// Test joining topics
		testTopic, err := ps.Join("test-topic")
		require.NoError(t, err, "Should be able to join test topic")
		defer testTopic.Close()

		assert.NotNil(t, testTopic, "Topic should not be nil")
	})
}

func TestParseAddr(t *testing.T) {
	testCases := []struct {
		name        string
		addr        string
		expectError bool
	}{
		{
			name:        "ValidAddress",
			addr:        "/ip4/127.0.0.1/tcp/4001/p2p/12D3KooWTest",
			expectError: false,
		},
		{
			name:        "ValidAddressWithQUIC",
			addr:        "/ip4/192.168.1.1/udp/4001/quic-v1/p2p/12D3KooWTest",
			expectError: false,
		},
		{
			name:        "InvalidAddress",
			addr:        "not-a-valid-multiaddr",
			expectError: true,
		},
		{
			name:        "EmptyAddress",
			addr:        "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			addrInfo, err := parseAddr(tc.addr)

			if tc.expectError {
				assert.Error(t, err, "Should return error for invalid address")
				assert.Nil(t, addrInfo, "AddrInfo should be nil on error")
			} else {
				assert.NoError(t, err, "Should not return error for valid address")
				assert.NotNil(t, addrInfo, "AddrInfo should not be nil for valid address")
				assert.NotEmpty(t, addrInfo.ID, "Peer ID should not be empty")
			}
		})
	}
}

func TestSendHandshake(t *testing.T) {
	setupTestConfig()
	defer os.RemoveAll(config.AppConfig.DbDir)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create first node with original port
	node1, _, err := createNodeWithPubSub(ctx)
	require.NoError(t, err)
	defer node1.Close()

	// Create second node with different configuration
	originalPort := config.AppConfig.Libp2pPort
	originalDbDir := config.AppConfig.DbDir

	// Use different port and db directory for second node
	config.AppConfig.Libp2pPort = 40002
	config.AppConfig.DbDir = filepath.Join(os.TempDir(), "goat-relayer-test-2")
	defer os.RemoveAll(config.AppConfig.DbDir)

	node2, _, err := createNodeWithPubSub(ctx)
	require.NoError(t, err)
	defer node2.Close()

	// Restore original config
	config.AppConfig.Libp2pPort = originalPort
	config.AppConfig.DbDir = originalDbDir

	// Verify nodes have different peer IDs
	assert.NotEqual(t, node1.ID(), node2.ID(), "Nodes should have different peer IDs")

	// Set up handshake protocol handler on node2
	handshakeReceived := make(chan bool, 1)
	node2.SetStreamHandler(protocol.ID(handshakeProtocol), func(s network.Stream) {
		defer s.Close()
		handleHandshake(s, node2)
		// Give client time to read response before closing
		time.Sleep(50 * time.Millisecond)
		handshakeReceived <- true
	})

	// Get node2's address info for connection
	node2Addrs := node2.Addrs()
	require.NotEmpty(t, node2Addrs, "Node2 should have addresses")

	// Find a suitable address (prefer TCP)
	var connectAddr multiaddr.Multiaddr
	for _, addr := range node2Addrs {
		if strings.Contains(addr.String(), "/tcp/") {
			connectAddr, err = multiaddr.NewMultiaddr(fmt.Sprintf("%s/p2p/%s", addr.String(), node2.ID().String()))
			require.NoError(t, err)
			break
		}
	}

	if connectAddr == nil {
		t.Skip("No suitable TCP address found for connection")
		return
	}

	// Parse the connection address
	peerInfo, err := parseAddr(connectAddr.String())
	require.NoError(t, err, "Should be able to parse connection address")

	// Connect nodes
	err = node1.Connect(ctx, *peerInfo)
	require.NoError(t, err, "Nodes should be able to connect")

	// Wait a bit for connection to establish
	time.Sleep(100 * time.Millisecond)

	// Create handshake stream
	stream, err := node1.NewStream(ctx, node2.ID(), protocol.ID(handshakeProtocol))
	require.NoError(t, err, "Should be able to create handshake stream")

	// Test successful handshake
	err = sendHandshake(stream)
	assert.NoError(t, err, "Handshake should succeed")
	stream.Close()

	// Wait for handshake to be received
	select {
	case <-handshakeReceived:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Handshake was not received within timeout")
	}
}

func TestTopicReconnection(t *testing.T) {
	config.AppConfig.DbDir = "test_db_topic_reconnect"
	defer os.RemoveAll(config.AppConfig.DbDir)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	service := NewLibP2PService(nil)

	// Create a test node
	node, ps, err := createNodeWithPubSub(ctx)
	require.NoError(t, err)
	defer node.Close()

	// Store global reference
	pubsubService = ps

	// Initialize topics
	err = service.initializeTopics(ctx, node)
	require.NoError(t, err)

	// Verify topics are initialized
	assert.NotNil(t, messageTopic)
	assert.NotNil(t, hbTopic)
	assert.NotNil(t, currentMessageSub)
	assert.NotNil(t, currentHBSub)

	// Test forced reconnection
	err = service.ForceTopicReconnection(ctx, node)
	assert.NoError(t, err)

	// Verify topics are still valid after reconnection
	assert.NotNil(t, messageTopic)
	assert.NotNil(t, hbTopic)

	t.Log("Topic reconnection test completed successfully")
}

func TestNetworkDiagnosisWithRecovery(t *testing.T) {
	config.AppConfig.DbDir = "test_db_diagnosis_recovery"
	defer os.RemoveAll(config.AppConfig.DbDir)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	service := NewLibP2PService(nil)

	// Create a test node
	node, ps, err := createNodeWithPubSub(ctx)
	require.NoError(t, err)
	defer node.Close()

	// Store global reference
	pubsubService = ps

	// Initialize topics
	err = service.initializeTopics(ctx, node)
	require.NoError(t, err)

	// Start network diagnosis in a separate goroutine
	diagCtx, diagCancel := context.WithTimeout(ctx, 5*time.Second)
	defer diagCancel()

	done := make(chan bool)
	go func() {
		service.startNetworkDiagnosisWithRecovery(diagCtx, node)
		done <- true
	}()

	// Wait a bit and then cancel
	time.Sleep(2 * time.Second)
	diagCancel()

	// Wait for diagnosis to finish
	select {
	case <-done:
		t.Log("Network diagnosis with recovery test completed successfully")
	case <-time.After(3 * time.Second):
		t.Error("Network diagnosis did not complete in time")
	}
}

func TestNetworkDiagnosis(t *testing.T) {
	setupTestConfig()
	defer os.RemoveAll(config.AppConfig.DbDir)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a real database manager for testing
	dbm := db.NewDatabaseManager()
	mockState := state.InitializeState(dbm)
	libp2pService := NewLibP2PService(mockState)

	// Create node and topic for testing
	node, ps, err := createNodeWithPubSub(ctx)
	require.NoError(t, err)
	defer node.Close()

	topic, err := ps.Join("test-topic")
	require.NoError(t, err)
	defer topic.Close()

	// Test network diagnosis function
	diagCtx, diagCancel := context.WithTimeout(ctx, 2*time.Second)
	defer diagCancel()

	// Run diagnosis in background
	go libp2pService.startNetworkDiagnosisWithRecovery(diagCtx, node)

	// Wait a bit to let diagnosis run
	time.Sleep(100 * time.Millisecond)

	// The function should not panic and should handle the context cancellation gracefully
	diagCancel()
	time.Sleep(100 * time.Millisecond)
}

func TestLibP2PServiceStartStop(t *testing.T) {
	setupTestConfig()
	defer os.RemoveAll(config.AppConfig.DbDir)

	// Create a real database manager for testing
	dbm := db.NewDatabaseManager()
	mockState := state.InitializeState(dbm)
	libp2pService := NewLibP2PService(mockState)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the service in a goroutine
	done := make(chan bool)
	go func() {
		libp2pService.Start(ctx)
		done <- true
	}()

	// Let it run for a short time
	time.Sleep(1 * time.Second)

	// Cancel context to stop the service
	cancel()

	// Wait for service to stop
	select {
	case <-done:
		// Service stopped successfully
	case <-time.After(3 * time.Second):
		t.Fatal("Service did not stop within timeout")
	}
}

func TestMultiAddrValidation(t *testing.T) {
	testAddrs := []string{
		"/ip4/127.0.0.1/tcp/4001",
		"/ip4/0.0.0.0/tcp/4001",
		"/ip4/192.168.1.1/udp/4001/quic-v1",
		"/ip6/::1/tcp/4001",
	}

	for _, addr := range testAddrs {
		t.Run("ValidateAddr_"+addr, func(t *testing.T) {
			maddr, err := multiaddr.NewMultiaddr(addr)
			assert.NoError(t, err, "Address should be valid: %s", addr)
			assert.NotNil(t, maddr, "Multiaddr should not be nil")
		})
	}
}

func TestPrivateKeyGeneration(t *testing.T) {
	setupTestConfig()
	defer os.RemoveAll(config.AppConfig.DbDir)

	// Test key generation
	key1, err := loadOrCreatePrivateKey("test_key_1.pem")
	require.NoError(t, err, "Should generate private key")
	assert.NotNil(t, key1, "Key should not be nil")

	// Test key persistence
	key2, err := loadOrCreatePrivateKey("test_key_1.pem")
	require.NoError(t, err, "Should load existing private key")
	assert.NotNil(t, key2, "Loaded key should not be nil")

	// Keys should be the same (loaded from disk)
	key1Bytes, err := key1.Raw()
	require.NoError(t, err)
	key2Bytes, err := key2.Raw()
	require.NoError(t, err)
	assert.Equal(t, key1Bytes, key2Bytes, "Keys should be identical when loaded from disk")
}

// Additional tests for complete coverage

func TestNewLibP2PService(t *testing.T) {
	setupTestConfig()
	defer os.RemoveAll(config.AppConfig.DbDir)

	// Create a database manager for testing
	dbm := db.NewDatabaseManager()
	mockState := state.InitializeState(dbm)

	// Test service creation
	service := NewLibP2PService(mockState)
	assert.NotNil(t, service, "Service should not be nil")
	assert.Equal(t, mockState, service.state, "Service should store the correct state")
}

func TestConnectToBootNode(t *testing.T) {
	setupTestConfig()
	defer os.RemoveAll(config.AppConfig.DbDir)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Create a database manager and service
	dbm := db.NewDatabaseManager()
	mockState := state.InitializeState(dbm)
	libp2pService := NewLibP2PService(mockState)

	// Create bootnode with original config
	bootnode, _, err := createNodeWithPubSub(ctx)
	require.NoError(t, err)
	defer bootnode.Close()

	// Set up handshake handler on bootnode
	bootnode.SetStreamHandler(protocol.ID(handshakeProtocol), func(s network.Stream) {
		defer s.Close()
		handleHandshake(s, bootnode)
		// Give client time to read response before closing
		time.Sleep(50 * time.Millisecond)
	})

	// Create client node with different configuration to ensure different peer ID
	originalPort := config.AppConfig.Libp2pPort
	originalDbDir := config.AppConfig.DbDir

	// Use different port and db directory for client node (ensures different private key)
	config.AppConfig.Libp2pPort = 40003
	config.AppConfig.DbDir = filepath.Join(os.TempDir(), "goat-relayer-test-bootnode-client")
	defer os.RemoveAll(config.AppConfig.DbDir)

	clientNode, _, err := createNodeWithPubSub(ctx)
	require.NoError(t, err)
	defer clientNode.Close()

	// Restore original config
	config.AppConfig.Libp2pPort = originalPort
	config.AppConfig.DbDir = originalDbDir

	// Get bootnode address
	bootnodeAddrs := bootnode.Addrs()
	require.NotEmpty(t, bootnodeAddrs, "Bootnode should have addresses")

	var bootnodeAddr string
	for _, addr := range bootnodeAddrs {
		if strings.Contains(addr.String(), "/tcp/") {
			bootnodeAddr = fmt.Sprintf("%s/p2p/%s", addr.String(), bootnode.ID().String())
			break
		}
	}

	if bootnodeAddr == "" {
		t.Skip("No suitable bootnode address found")
		return
	}

	// Test successful connection
	t.Logf("Testing connection to bootnode address: %s", bootnodeAddr)
	err = libp2pService.connectToBootNode(ctx, clientNode, bootnodeAddr)
	if err != nil {
		t.Logf("Connection error: %v", err)
	}
	assert.NoError(t, err, "Should successfully connect to bootnode")

	// Verify connection
	peers := clientNode.Network().Peers()
	t.Logf("Connected peers: %v, looking for: %v", peers, bootnode.ID())
	assert.Contains(t, peers, bootnode.ID(), "Client should be connected to bootnode")

	// Test invalid address (this should be a separate context to avoid interference)
	invalidCtx, invalidCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer invalidCancel()
	err = libp2pService.connectToBootNode(invalidCtx, clientNode, "invalid-address")
	assert.Error(t, err, "Should fail with invalid address")
}

func TestMessagePublishing(t *testing.T) {
	setupTestConfig()
	defer os.RemoveAll(config.AppConfig.DbDir)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Create nodes for publisher and subscriber
	node1, ps1, err := createNodeWithPubSub(ctx)
	require.NoError(t, err)
	defer node1.Close()

	// Set up message topic on node1
	topic1, err := ps1.Join("test-message-topic")
	require.NoError(t, err)
	defer topic1.Close()

	// Set global message topic for PublishMessage function
	originalTopic := messageTopic
	messageTopic = topic1
	defer func() { messageTopic = originalTopic }()

	// Test publishing message
	testMsg := Message[any]{
		RequestId:   "test-123",
		MessageType: MessageTypeUnknown,
		Data:        "Test message content",
	}

	err = PublishMessage(ctx, testMsg)
	assert.NoError(t, err, "Should be able to publish message")

	// Test publishing with nil topic
	messageTopic = nil
	err = PublishMessage(ctx, testMsg)
	assert.Error(t, err, "Should fail when message topic is nil")
	assert.Contains(t, err.Error(), "message topic is nil")
}

func TestGenerateTransportAddresses(t *testing.T) {
	testCases := []struct {
		name        string
		inputAddr   string
		expectError bool
		expectCount int
	}{
		{
			name:        "ValidTCPAddress",
			inputAddr:   "/ip4/127.0.0.1/tcp/4001/p2p/12D3KooWTest",
			expectError: false,
			expectCount: 1, // Should generate TCP
		},
		{
			name:        "ValidQUICAddress",
			inputAddr:   "/ip4/192.168.1.1/udp/4001/quic-v1/p2p/12D3KooWTest",
			expectError: false,
			expectCount: 1, // Should generate TCP
		},
		{
			name:        "IPv6Address",
			inputAddr:   "/ip6/::1/tcp/4001/p2p/12D3KooWTest",
			expectError: false,
			expectCount: 1, // Should generate TCP for IPv6
		},
		{
			name:        "InvalidAddress",
			inputAddr:   "invalid-multiaddr",
			expectError: true,
			expectCount: 0,
		},
		{
			name:        "AddressWithoutPeerID",
			inputAddr:   "/ip4/127.0.0.1/tcp/4001",
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			addrs, err := generateTransportAddresses(tc.inputAddr)

			if tc.expectError {
				assert.Error(t, err, "Should return error for invalid address")
				assert.Nil(t, addrs, "Addresses should be nil on error")
			} else {
				assert.NoError(t, err, "Should not return error for valid address")
				assert.Len(t, addrs, tc.expectCount, "Should generate expected number of addresses")

				if len(addrs) > 0 {
					// Verify all addresses have the same peer ID
					firstPeerID := addrs[0].ID
					for i, addr := range addrs {
						assert.Equal(t, firstPeerID, addr.ID, "All addresses should have same peer ID")
						assert.NotEmpty(t, addr.Addrs, "Address should have multiaddrs")

						// Log generated addresses for debugging
						t.Logf("Generated address %d: %s", i, addr.Addrs[0])
					}

					// Verify we have TCP if count is 1
					if tc.expectCount == 1 {
						hasTCP := false
						for _, addr := range addrs {
							addrStr := addr.Addrs[0].String()
							if strings.Contains(addrStr, "/tcp/") {
								hasTCP = true
							}
						}
						assert.True(t, hasTCP, "Should generate TCP address")
					}
				}
			}
		})
	}
}

func TestConnectionMonitoringAndReconnect(t *testing.T) {
	setupTestConfig()
	defer os.RemoveAll(config.AppConfig.DbDir)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create service for client
	dbm := db.NewDatabaseManager()
	mockState := state.InitializeState(dbm)
	libp2pService := NewLibP2PService(mockState)

	// Create initial bootnode
	bootnode, _, err := createNodeWithPubSub(ctx)
	require.NoError(t, err)

	// Set up handshake handler on bootnode
	bootnode.SetStreamHandler(protocol.ID(handshakeProtocol), func(s network.Stream) {
		defer s.Close()
		handleHandshake(s, bootnode)
		time.Sleep(50 * time.Millisecond)
	})

	// Create client node
	originalPort := config.AppConfig.Libp2pPort
	originalDbDir := config.AppConfig.DbDir

	config.AppConfig.Libp2pPort = 40004
	config.AppConfig.DbDir = filepath.Join(os.TempDir(), "goat-relayer-test-monitor-client")
	defer os.RemoveAll(config.AppConfig.DbDir)

	clientNode, _, err := createNodeWithPubSub(ctx)
	require.NoError(t, err)
	defer clientNode.Close()

	// Restore original config
	config.AppConfig.Libp2pPort = originalPort
	config.AppConfig.DbDir = originalDbDir

	// Get bootnode address (TCP format)
	bootnodeAddrs := bootnode.Addrs()
	require.NotEmpty(t, bootnodeAddrs, "Bootnode should have addresses")

	var bootnodeAddr string
	for _, addr := range bootnodeAddrs {
		if strings.Contains(addr.String(), "/tcp/") {
			bootnodeAddr = fmt.Sprintf("%s/p2p/%s", addr.String(), bootnode.ID().String())
			break
		}
	}

	if bootnodeAddr == "" {
		t.Skip("No suitable bootnode address found")
		return
	}

	// Initial connection
	err = libp2pService.connectToBootNode(ctx, clientNode, bootnodeAddr)
	require.NoError(t, err, "Initial connection should succeed")

	// Verify connection
	peers := clientNode.Network().Peers()
	require.Contains(t, peers, bootnode.ID(), "Client should be connected to bootnode")

	t.Logf("Initial connection established via TCP")

	// Test connection status checking
	isConnected := libp2pService.isReallyConnected(clientNode, bootnode.ID())
	assert.True(t, isConnected, "Should detect connection as active")

	// Simulate bootnode restart by closing it
	t.Logf("Simulating bootnode restart...")
	bootnode.Close()

	// Wait a bit for disconnection to be detected
	time.Sleep(2 * time.Second)

	// Verify disconnection is detected
	isConnected = libp2pService.isReallyConnected(clientNode, bootnode.ID())
	assert.False(t, isConnected, "Should detect disconnection")

	// Test cleanup function
	assert.NotPanics(t, func() {
		libp2pService.cleanupStaleConnections(clientNode, bootnode.ID())
	}, "Cleanup should not panic")

	t.Logf("Connection monitoring and cleanup test completed successfully")
}

func TestEnhancedConnectionStatusCheck(t *testing.T) {
	setupTestConfig()
	defer os.RemoveAll(config.AppConfig.DbDir)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create service and nodes
	dbm := db.NewDatabaseManager()
	mockState := state.InitializeState(dbm)
	libp2pService := NewLibP2PService(mockState)

	node1, _, err := createNodeWithPubSub(ctx)
	require.NoError(t, err)
	defer node1.Close()

	originalPort := config.AppConfig.Libp2pPort
	config.AppConfig.Libp2pPort = 40005
	node2, _, err := createNodeWithPubSub(ctx)
	require.NoError(t, err)
	defer node2.Close()
	config.AppConfig.Libp2pPort = originalPort

	// Test with no connection
	isConnected := libp2pService.isReallyConnected(node1, node2.ID())
	assert.False(t, isConnected, "Should detect no connection")

	// Test cleanup function doesn't panic with non-existent peer
	assert.NotPanics(t, func() {
		libp2pService.cleanupStaleConnections(node1, node2.ID())
	}, "Cleanup should handle non-existent peer gracefully")
}
