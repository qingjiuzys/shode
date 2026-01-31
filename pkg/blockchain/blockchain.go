// Package blockchain 提供区块链功能。
package blockchain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Blockchain 区块链
type Blockchain struct {
	blocks       []*Block
	transactions []*Transaction
	pendingTxs   []*Transaction
	difficulty   int
	reward       float64
	mu           sync.RWMutex
}

// Block 区块
type Block struct {
	Index        int            `json:"index"`
	Timestamp    time.Time      `json:"timestamp"`
	Transactions []*Transaction `json:"transactions"`
	PrevHash     string         `json:"prev_hash"`
	Hash         string         `json:"hash"`
	Nonce        int            `json:"nonce"`
}

// Transaction 交易
type Transaction struct {
	ID        string                 `json:"id"`
	Sender    string                 `json:"sender"`
	Receiver  string                 `json:"receiver"`
	Amount    float64                `json:"amount"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Signature string                 `json:"signature"`
	Status    string                 `json:"status"`
}

// SmartContract 智能合约
type SmartContract struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Bytecode    []byte                 `json:"bytecode"`
	State       map[string]interface{} `json:"state"`
	ABI         *ABI                   `json:"abi"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ABI 应用二进制接口
type ABI struct {
	Methods []ABIMethod `json:"methods"`
	Events  []ABIEvent  `json:"events"`
}

// ABIMethod ABI 方法
type ABIMethod struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Inputs     []ABIInput             `json:"inputs"`
	Outputs    []ABIOutput            `json:"outputs"`
	StateMutability bool             `json:"stateMutability"`
}

// ABIInput ABI 输入
type ABIInput struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ABIOutput ABI 输出
type ABIOutput struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ABIEvent ABI 事件
type ABIEvent struct {
	Name   string     `json:"name"`
	Inputs []ABIInput `json:"inputs"`
}

// NewBlockchain 创建区块链
func NewBlockchain() *Blockchain {
	genesis := createGenesisBlock()
	return &Blockchain{
		blocks:     []*Block{genesis},
		difficulty: 4,
		reward:     10.0,
	}
}

// createGenesisBlock 创世区块
func createGenesisBlock() *Block {
	return &Block{
		Index:     0,
		Timestamp: time.Now(),
		PrevHash:  "0",
		Hash:      calculateBlockHash(0, time.Now(), []*Transaction{}, "0", 0),
	}
}

// AddTransaction 添加交易
func (bc *Blockchain) AddTransaction(tx *Transaction) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	tx.ID = calculateTxHash(tx)
	tx.Status = "pending"
	tx.Timestamp = time.Now()

	bc.pendingTxs = append(bc.pendingTxs, tx)

	return nil
}

// Mine 挖矿
func (bc *Blockchain) Mine(minerAddress string) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// 选择交易
	txs := bc.selectTransactions()

	// 创建新区块
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := &Block{
		Index:        prevBlock.Index + 1,
		Timestamp:    time.Now(),
		Transactions: txs,
		PrevHash:     prevBlock.Hash,
	}

	// 工作量证明
	nonce := bc.proofOfWork(newBlock)

	newBlock.Nonce = nonce
	newBlock.Hash = calculateBlockHash(newBlock.Index, newBlock.Timestamp, txs, newBlock.PrevHash, nonce)

	// 挖矿奖励
	rewardTx := &Transaction{
		Sender:   "0x0000000000000000000000000000000000000000",
		Receiver: minerAddress,
		Amount:   bc.reward,
		Timestamp: time.Now(),
	}
	rewardTx.ID = calculateTxHash(rewardTx)
	rewardTx.Status = "confirmed"

	newBlock.Transactions = append([]*Transaction{rewardTx}, txs...)

	bc.blocks = append(bc.blocks, newBlock)

	// 清理待处理交易
	bc.clearPendingTransactions(txs)

	return nil
}

// selectTransactions 选择交易
func (bc *Blockchain) selectTransactions() []*Transaction {
	maxTxs := 10
	if len(bc.pendingTxs) < maxTxs {
		maxTxs = len(bc.pendingTxs)
	}

	selected := make([]*Transaction, maxTxs)
	copy(selected, bc.pendingTxs[:maxTxs])

	return selected
}

// clearPendingTransactions 清理待处理交易
func (bc *Blockchain) clearPendingTransactions(txs []*Transaction) {
	newPending := make([]*Transaction, 0)
	txMap := make(map[string]bool)

	for _, tx := range txs {
		txMap[tx.ID] = true
	}

	for _, tx := range bc.pendingTxs {
		if !txMap[tx.ID] {
			newPending = append(newPending, tx)
		}
	}

	bc.pendingTxs = newPending
}

// proofOfWork 工作量证明
func (bc *Blockchain) proofOfWork(block *Block) int {
	var hash string
	var nonce int

	target := fmt.Sprintf("%0"+bc.difficulty+"s", 0)

	for {
		hash = calculateBlockHash(block.Index, block.Timestamp, block.Transactions, block.PrevHash, nonce)
		if hash[:len(target)] == target {
			break
		}
		nonce++
	}

	return nonce
}

// GetBalance 获取余额
func (bc *Blockchain) GetBalance(address string) float64 {
	bc.mu.RLock()
	defer bc.RUnlock()

	balance := 0.0

	for _, block := range bc.blocks {
		for _, tx := range block.Transactions {
			if tx.Sender == address {
				balance -= tx.Amount
			}
			if tx.Receiver == address {
				balance += tx.Amount
			}
		}
	}

	return balance
}

// GetBlock 获取区块
func (bc *Blockchain) GetBlock(index int) (*Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if index < 0 || index >= len(bc.blocks) {
		return nil, fmt.Errorf("block not found: %d", index)
	}

	return bc.blocks[index], nil
}

// GetChainLength 获取链长度
func (bc *Blockchain) GetChainLength() int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return len(bc.blocks)
}

// VerifyChain 验证链
func (bc *Blockchain) VerifyChain() bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for i := 1; i < len(bc.blocks); i++ {
		current := bc.blocks[i]
		prev := bc.blocks[i-1]

		if current.PrevHash != prev.Hash {
			return false
		}

		// 验证哈希
		calculatedHash := calculateBlockHash(current.Index, current.Timestamp, current.Transactions, current.PrevHash, current.Nonce)
		if calculatedHash != current.Hash {
			return false
		}
	}

	return true
}

// SmartContractManager 智能合约管理器
type SmartContractManager struct {
	contracts map[string]*SmartContract
	mu        sync.RWMutex
}

// NewSmartContractManager 创建智能合约管理器
func NewSmartContractManager() *SmartContractManager {
	return &SmartContractManager{
		contracts: make(map[string]*SmartContract),
	}
}

// Deploy 部署合约
func (scm *SmartContractManager) Deploy(name string, bytecode []byte, abi *ABI) (*SmartContract, error) {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	contract := &SmartContract{
		ID:        generateContractID(),
		Name:      name,
		Bytecode:  bytecode,
		State:     make(map[string]interface{}),
		ABI:       abi,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	scm.contracts[contract.ID] = contract

	return contract, nil
}

// Call 调用合约方法
func (scm *SmartContractManager) Call(ctx context.Context, contractID, method string, args map[string]interface{}) (interface{}, error) {
	scm.mu.RLock()
	defer scm.mu.RUnlock()

	contract, exists := scm.contracts[contractID]
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractID)
	}

	// 查找方法
	var abimethod *ABIMethod
	for _, m := range contract.ABI.Methods {
		if m.Name == method {
			abimethod = &m
			break
		}
	}

	if abimethod == nil {
		return nil, fmt.Errorf("method not found: %s", method)
	}

	// 简化实现，返回固定值
	return fmt.Sprintf("called %s on contract %s with %v", method, contract.Name, args), nil
}

// GetState 获取合约状态
func (scm *SmartContractManager) GetState(contractID string) (map[string]interface{}, error) {
	scm.mu.RLock()
	defer scm.mu.RUnlock()

	contract, exists := scm.contracts[contractID]
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractID)
	}

	return contract.State, nil
}

// SetState 设置合约状态
func (scm *SmartContractManager) SetState(contractID string, key string, value interface{}) error {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	contract, exists := scm.contracts[contractID]
	if !exists {
		return fmt.Errorf("contract not found: %s", contractID)
	}

	contract.State[key] = value
	contract.UpdatedAt = time.Now()

	return nil
}

// NFTManager NFT 管理器
type NFTManager struct {
	tokens map[string]*NFT
	mu     sync.RWMutex
}

// NFT NFT
type NFT struct {
	TokenID     string `json:"token_id"`
	ContractID   string `json:"contract_id"`
	Owner       string `json:"owner"`
	Metadata    *NFTMetadata `json:"metadata"`
	CreatedAt   time.Time `json:"created_at"`
}

// NFTMetadata NFT 元数据
type NFTMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Attributes  map[string]string `json:"attributes"`
}

// NewNFTManager 创建 NFT 管理器
func NewNFTManager() *NFTManager {
	return &NFTManager{
		tokens: make(map[string]*NFT),
	}
}

// Mint 铸造 NFT
func (nm *NFTManager) Mint(contractID, owner string, metadata *NFTMetadata) (*NFT, error) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	tokenID := generateNFTID()

	nft := &NFT{
		TokenID:   tokenID,
		ContractID: contractID,
		Owner:     owner,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}

	nm.tokens[tokenID] = nft

	return nft, nil
}

// Transfer 转移 NFT
func (nm *NFTManager) Transfer(tokenID, from, to string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nft, exists := nm.tokens[tokenID]
	if !exists {
		return fmt.Errorf("NFT not found: %s", tokenID)
	}

	if nft.Owner != from {
		return fmt.Errorf("not owner of NFT: %s", tokenID)
	}

	nft.Owner = to

	return nil
}

// GetToken 获取 NFT
func (nm *NFTManager) GetToken(tokenID string) (*NFT, bool) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	token, exists := nm.tokens[tokenID]
	return token, exists
}

// DIDManager 去中心化身份管理器
type DIDManager struct {
	identities map[string]*DIDDocument
	mu         sync.RWMutex
}

// DIDDocument DID 文档
type DIDDocument struct {
	ID              string                 `json:"id"`
	Context         string                 `json:"@context"`
	PublicKey       []PublicKey           `json:"publicKey"`
	Authentication  []AuthenticationMethod  `json:"authentication"`
	Service         []Service              `json:"service"`
	Created         time.Time              `json:"created"`
	Updated         time.Time              `json:"updated"`
}

// PublicKey 公钥
type PublicKey struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	PublicKey string `json:"publicKey"`
}

// AuthenticationMethod 认证方法
type AuthenticationMethod struct {
	Type          string `json:"type"`
	PublicKeyID   string `json:"publicKeyID"`
}

// Service 服务
type Service struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

// NewDIDManager 创建 DID 管理器
func NewDIDManager() *DIDManager {
	return &DIDManager{
		identities: make(map[string]*DIDDocument),
	}
}

// Create 创建 DID
func (dm *DIDManager) Create(did string, publicKey string) (*DIDDocument, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	doc := &DIDDocument{
		ID:        did,
		Context:   "https://w3id.org/did/v1",
		PublicKey: []PublicKey{
			{
				ID:        did + "#keys-1",
				Type:      "Secp256k1",
				PublicKey: publicKey,
			},
		},
		Authentication: []AuthenticationMethod{
			{
				Type:        "Secp256k1",
				PublicKeyID: did + "#keys-1",
			},
		},
		Created: time.Now(),
		Updated: time.Now(),
	}

	dm.identities[did] = doc

	return doc, nil
}

// Resolve 解析 DID
func (dm *DIDManager) Resolve(did string) (*DIDDocument, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	doc, exists := dm.identities[did]
	return doc, exists
}

// Verify 验证 DID
func (dm *DIDManager) Verify(did string, signature, message string) (bool, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	doc, exists := dm.identities[did]
	if !exists {
		return false, fmt.Errorf("DID not found: %s", did)
	}

	// 简化实现，总是返回 true
	return true, nil
}

// calculateBlockHash 计算区块哈希
func calculateBlockHash(index int, timestamp time.Time, txs []*Transaction, prevHash string, nonce int) string {
	data := fmt.Sprintf("%d%d%s%d", index, timestamp.UnixNano(), prevHash, nonce)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash)
}

// calculateTxHash 计算交易哈希
func calculateTxHash(tx *Transaction) string {
	data := fmt.Sprintf("%s%s%f%d", tx.Sender, tx.Receiver, tx.Amount, tx.Timestamp.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash)[:16]
}

// generateContractID 生成合约 ID
func generateContractID() string {
	return fmt.Sprintf("0x%x", sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano()))))
}

// generateNFTID 生成 NFT ID
func generateNFTID() string {
	return fmt.Sprintf("nft_%d", time.Now().UnixNano())
}

// LedgerData 链上数据存证
type LedgerData struct {
	ID        string                 `json:"id"`
	Data      interface{}            `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Hash      string                 `json:"hash"`
	Signature string                 `json:"signature"`
	StoreTime time.Time              `json:"store_time"`
}

// Ledger 链上账本
type Ledger struct {
	data  map[string]*LedgerData
	blockchain *Blockchain
	mu    sync.RWMutex
}

// NewLedger 创建账本
func NewLedger(blockchain *Blockchain) *Ledger {
	return &Ledger{
		data:       make(map[string]*LedgerData),
		blockchain: blockchain,
	}
}

// Store 存储数据上链
func (l *Ledger) Store(data interface{}) (*LedgerData, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 序列化数据
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(bytes)
	id := hex.EncodeToString(hash)[:16]

	ledgerData := &LedgerData{
		ID:        id,
		Data:      data,
		Timestamp: time.Now(),
		Hash:      hex.EncodeToString(hash),
		StoreTime: time.Now(),
	}

	l.data[id] = ledgerData

	return ledgerData, nil
}

// Retrieve 检索数据
func (l *Ledger) Retrieve(id string) (*LedgerData, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	data, exists := l.data[id]
	return data, exists
}

// Verify 验证数据
func (l *Ledger) Verify(id string) (bool, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	data, exists := l.data[id]
	if !exists {
		return false, fmt.Errorf("data not found: %s", id)
	}

	// 验证哈希
	bytes, err := json.Marshal(data.Data)
	if err != nil {
		return false, err
	}

	hash := sha256.Sum256(bytes)
	calculatedHash := hex.EncodeToString(hash)

	return calculatedHash == data.Hash, nil
}

// CryptoPayment 加密货币支付
type CryptoPayment struct {
	ID          string                 `json:"id"`
	From        string                 `json:"from"`
	To          string                 `json:"to"`
	Amount      float64                `json:"amount"`
	Currency    string                 `json:"currency"`
	Timestamp   time.Time              `json:"timestamp"`
	Status      string                 `json:"status"`
	TransactionHash string           `json:"transaction_hash"`
	Fee         float64                `json:"fee"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PaymentProcessor 支付处理器
type PaymentProcessor struct {
	payments map[string]*CryptoPayment
	mu       sync.RWMutex
}

// NewPaymentProcessor 创建支付处理器
func NewPaymentProcessor() *PaymentProcessor {
	return &PaymentProcessor{
		payments: make(map[string]*CryptoPayment),
	}
}

// CreatePayment 创建支付
func (pp *PaymentProcessor) CreatePayment(from, to string, amount float64, currency string) (*CryptoPayment, error) {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	payment := &CryptoPayment{
		ID:        generatePaymentID(),
		From:      from,
		To:        to,
		Amount:    amount,
		Currency:  currency,
		Timestamp: time.Now(),
		Status:    "pending",
	}

	pp.payments[payment.ID] = payment

	return payment, nil
}

// ProcessPayment 处理支付
func (pp *PaymentProcessor) ProcessPayment(paymentID string) error {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	payment, exists := pp.payments[paymentID]
	if !exists {
		return fmt.Errorf("payment not found: %s", paymentID)
	}

	// 简化实现，直接标记为完成
	payment.Status = "completed"

	return nil
}

// GetPayment 获取支付
func (pp *PaymentProcessor) GetPayment(paymentID string) (*CryptoPayment, bool) {
	pp.mu.RLock()
	defer pp.mu.RUnlock()

	payment, exists := pp.payments[paymentID]
	return payment, exists
}

// generatePaymentID 生成支付 ID
func generatePaymentID() string {
	return fmt.Sprintf("pay_%d", time.Now().UnixNano())
}
