// Package webrtc WebRTC P2P 通信
package webrtc

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PeerConnection WebRTC 对等连接
type PeerConnection struct {
	ID         string
	LocalID    string
	RemoteID   string
	State      string
	DataChannels map[string]*DataChannel
	mu         sync.RWMutex
}

// DataChannel 数据通道
type DataChannel struct {
	ID      string
	Label   string
	SendCh  chan []byte
	RecvCh  chan []byte
	mu      sync.RWMutex
}

// WebRTCManager WebRTC 管理器
type WebRTCManager struct {
	connections map[string]*PeerConnection
	offers      map[string]string
	answers     map[string]string
	mu          sync.RWMutex
}

// NewWebRTCManager 创建管理器
func NewWebRTCManager() *WebRTCManager {
	return &WebRTCManager{
		connections: make(map[string]*PeerConnection),
		offers:      make(map[string]string),
		answers:     make(map[string]string),
	}
}

// CreatePeerConnection 创建对等连接
func (wm *WebRTCManager) CreatePeerConnection(localID, remoteID string) (*PeerConnection, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	pc := &PeerConnection{
		ID:           generateID(),
		LocalID:      localID,
		RemoteID:     remoteID,
		State:        "new",
		DataChannels: make(map[string]*DataChannel),
	}

	wm.connections[pc.ID] = pc

	return pc, nil
}

// CreateOffer 创建 Offer
func (wm *WebRTCManager) CreateOffer(pc *PeerConnection) (string, error) {
	offer := generateSDP("offer")
	wm.mu.Lock()
	wm.offers[pc.ID] = offer
	wm.mu.Unlock()

	return offer, nil
}

// CreateAnswer 创建 Answer
func (wm *WebRTCManager) CreateAnswer(pc *PeerConnection, offer string) (string, error) {
	answer := generateSDP("answer")
	wm.mu.Lock()
	wm.answers[pc.ID] = answer
	wm.mu.Unlock()

	return answer, nil
}

// SetRemoteDescription 设置远程描述
func (wm *WebRTCManager) SetRemoteDescription(pc *PeerConnection, sdp string) error {
	pc.mu.Lock()
	pc.State = "connecting"
	pc.mu.Unlock()

	return nil
}

// SetLocalDescription 设置本地描述
func (wm *WebRTCManager) SetLocalDescription(pc *PeerConnection, sdp string) error {
	return nil
}

// AddICECandidate 添加 ICE 候选
func (wm *WebRTCManager) AddICECandidate(pc *PeerConnection, candidate string) error {
	return nil
}

// CreateDataChannel 创建数据通道
func (wm *WebRTCManager) CreateDataChannel(pc *PeerConnection, label string) (*DataChannel, error) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	dc := &DataChannel{
		ID:     generateID(),
		Label:  label,
		SendCh: make(chan []byte, 256),
		RecvCh: make(chan []byte, 256),
	}

	pc.DataChannels[label] = dc

	return dc, nil
}

// ClosePeerConnection 关闭对等连接
func (wm *WebRTCManager) ClosePeerConnection(pc *PeerConnection) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	delete(wm.connections, pc.ID)
	delete(wm.offers, pc.ID)
	delete(wm.answers, pc.ID)

	pc.mu.Lock()
	pc.State = "closed"

	// 关闭所有数据通道
	for _, dc := range pc.DataChannels {
		close(dc.SendCh)
		close(dc.RecvCh)
	}
	pc.mu.Unlock()

	return nil
}

// GetConnection 获取连接
func (wm *WebRTCManager) GetConnection(id string) (*PeerConnection, error) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if pc, ok := wm.connections[id]; ok {
		return pc, nil
	}
	return nil, fmt.Errorf("connection not found")
}

// SendData 发送数据
func (dc *DataChannel) SendData(data []byte) error {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	select {
	case dc.SendCh <- data:
		return nil
	default:
		return fmt.Errorf("channel full")
	}
}

// RecvData 接收数据
func (dc *DataChannel) RecvData() ([]byte, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	data, ok := <-dc.RecvCh
	if !ok {
		return nil, fmt.Errorf("channel closed")
	}
	return data, nil
}

// Close 关闭数据通道
func (dc *DataChannel) Close() error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	close(dc.SendCh)
	close(dc.RecvCh)

	return nil
}

// generateSDP 生成 SDP
func generateSDP(typ string) string {
	return fmt.Sprintf("v=0\r\no=- %d 0 IN IP4 0.0.0.0\r\ns=-\r\nt=0 0\r\na=%s",
		time.Now().Unix(), typ)
}

// generateID 生成唯一 ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Session WebRTC 会话
type Session struct {
	ID         string
	PeerConnection *PeerConnection
	Context    context.Context
	Cancel     context.CancelFunc
	mu         sync.RWMutex
}

// NewSession 创建会话
func NewSession(pc *PeerConnection) *Session {
	ctx, cancel := context.WithCancel(context.Background())

	return &Session{
		ID:            generateID(),
		PeerConnection: pc,
		Context:       ctx,
		Cancel:        cancel,
	}
}

// Close 关闭会话
func (s *Session) Close() error {
	s.Cancel()
	return nil
}

// Room WebRTC 房间
type Room struct {
	ID          string
	Clients     map[string]*PeerConnection
	mu          sync.RWMutex
}

// NewRoom 创建房间
func NewRoom(id string) *Room {
	return &Room{
		ID:      id,
		Clients: make(map[string]*PeerConnection),
	}
}

// Join 加入房间
func (r *Room) Join(clientID string, pc *PeerConnection) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Clients[clientID] = pc
}

// Leave 离开房间
func (r *Room) Leave(clientID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Clients, clientID)
}

// Broadcast 广播到房间
func (r *Room) Broadcast(data []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, pc := range r.Clients {
		pc.mu.RLock()
		for _, dc := range pc.DataChannels {
			select {
			case dc.SendCh <- data:
			default:
			}
		}
		pc.mu.RUnlock()
	}
}

// GetClientCount 获取客户端数量
func (r *Room) GetClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.Clients)
}
