// Package k8s Kubernetes 部署工具
package k8s

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// K8sDeployer Kubernetes 部署器
type K8sDeployer struct {
	config     *K8sConfig
	namespace  string
	dryRun     bool
	kubeconfig  string
}

// K8sConfig Kubernetes 配置
type K8sConfig struct {
	AppName        string
	Image         string
	Replicas      int
	ContainerPort int
	ServicePort   int
	EnvVars       map[string]string
	Resources     *K8sResources
	Limits        *K8sResources
	NodeSelector   map[string]string
	Affinity      *K8sAffinity
	Tolerations    []Toleration
}

// K8sResources Kubernetes 资源配置
type K8sResources struct {
	CPU    string
	Memory string
}

// K8sAffinity 亲和性配置
type K8sAffinity struct {
	NodeAffinity    map[string]string
	PodAffinity     map[string]string
}

// Toleration 容忍配置
type Toleration struct {
	Key      string
	Operator string
	Value    string
	Effect   string
}

// NewK8sDeployer 创建 Kubernetes 部署器
func NewK8sDeployer(config *K8sConfig) *K8sDeployer {
	return &K8sDeployer{
		config:    config,
		namespace: "default",
		dryRun:    false,
		kubeconfig: "",
	}
}

// Init 初始化 Kubernetes 项目
func (kd *K8sDeployer) Init() error {
	fmt.Println("☸️  Initializing Kubernetes project...")

	// 创建 k8s 目录
	if err := os.MkdirAll("k8s", 0755); err != nil {
		return fmt.Errorf("failed to create k8s directory: %w", err)
	}

	// 生成部署清单
	if err := kd.generateDeployment(); err != nil {
		return fmt.Errorf("failed to generate deployment: %w", err)
	}

	// 生成服务清单
	if err := kd.generateService(); err != nil {
		return fmt.Errorf("failed to generate service: %w", err)
	}

	// 生成 ConfigMap
	if err := kd.generateConfigMap(); err != nil {
		return fmt.Errorf("failed to generate configmap: %w", err)
	}

	// 生成 Secret
	if err := kd.generateSecret(); err != nil {
		return fmt.Errorf("failed to generate secret: %w", err)
	}

	// 生成 Ingress
	if err := kd.generateIngress(); err != nil {
		return fmt.Errorf("failed to generate ingress: %w", err)
	}

	// 生成 HPA
	if err := kd.generateHPA(); err != nil {
		return fmt.Errorf("failed to generate HPA: %w", err)
	}

	// 生成 Namespace
	if err := kd.generateNamespace(); err != nil {
		return fmt.Errorf("failed to generate namespace: %w", err)
	}

	fmt.Println("✓ Kubernetes project initialized")
	fmt.Println("\nGenerated files:")
	fmt.Println("  k8s/deployment.yaml")
	fmt.Println("  k8s/service.yaml")
	fmt.Println("  k8s/configmap.yaml")
	fmt.Println("  k8s/secret.yaml")
	fmt.Println("  k8s/ingress.yaml")
	fmt.Println("  k8s/hpa.yaml")
	fmt.Println("  k8s/namespace.yaml")
	fmt.Println("\nNext steps:")
	fmt.Println("  shode deploy k8s apply -f k8s/")
	fmt.Println("  shode deploy k8s get pods")
	fmt.Println("  shode deploy k8s get services")

	return nil
}

// generateDeployment 生成部署清单
func (kd *K8sDeployer) generateDeployment() error {
	deployment := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  namespace: %s
  labels:
    app: %s
spec:
  replicas: %d
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
      - name: %s
        image: %s
        ports:
        - containerPort: %d
        env:
        - name: ENV
          value: "production"
        %s
        resources:
          requests:
            cpu: "%s"
            memory: "%s"
          limits:
            cpu: "%s"
            memory: "%s"
        livenessProbe:
          httpGet:
            path: /health
            port: %d
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: %d
          initialDelaySeconds: 5
          periodSeconds: 5
      %s
      %s
`,
		kd.config.AppName,
		kd.namespace,
		kd.config.AppName,
		kd.config.Replicas,
		kd.config.AppName,
		kd.config.AppName,
		kd.config.AppName,
		kd.config.AppName,
		kd.config.Image,
		kd.config.ContainerPort,
		kd.generateEnvVars(),
		kd.config.Resources.CPU,
		kd.config.Resources.Memory,
		kd.config.Limits.CPU,
		kd.config.Limits.Memory,
		kd.config.ContainerPort,
		kd.config.ContainerPort,
		kd.generateNodeSelector(),
		kd.generateTolerations(),
	)

	return kd.writeManifest("deployment.yaml", deployment)
}

// generateService 生成服务清单
func (kd *K8sDeployer) generateService() error {
	service := fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
  namespace: %s
  labels:
    app: %s
spec:
  type: ClusterIP
  ports:
  - port: %d
    targetPort: %d
    protocol: TCP
  selector:
    app: %s
`,
		kd.config.AppName,
		kd.namespace,
		kd.config.AppName,
		kd.config.ServicePort,
		kd.config.ContainerPort,
		kd.config.AppName,
	)

	return kd.writeManifest("service.yaml", service)
}

// generateConfigMap 生成 ConfigMap
func (kd *K8sDeployer) generateConfigMap() error {
	configMap := fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: %s-config
  namespace: %s
data:
  app.shode: |
    server:
      port: 8080
    logging:
      level: info
`,
		kd.config.AppName,
		kd.namespace,
	)

	return kd.writeManifest("configmap.yaml", configMap)
}

// generateSecret 生成 Secret
func (kd *K8sDeployer) generateSecret() error {
	secret := fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: %s-secret
  namespace: %s
type: Opaque
stringData:
  # Base64 encoded values
  # Example: echo -n 'admin' | base64
  password: YWRtaW4=
  api-key: <your-api-key-here>
`,
		kd.config.AppName,
		kd.namespace,
	)

	return kd.writeManifest("secret.yaml", secret)
}

// generateIngress 生成 Ingress
func (kd *K8sDeployer) generateIngress() error {
	ingress := fmt.Sprintf(`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: %s-ingress
  namespace: %s
  annotations:
    kubernetes.io/ingress.class: "nginx"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - %s.example.com
    secretName: %s-tls
  rules:
  - host: %s.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: %s
            port:
              number: %d
`,
		kd.config.AppName,
		kd.namespace,
		kd.config.AppName,
		kd.config.AppName,
		kd.config.AppName,
		kd.config.AppName,
		kd.config.AppName,
		kd.config.ServicePort,
	)

	return kd.writeManifest("ingress.yaml", ingress)
}

// generateHPA 生成 HPA
func (kd *K8sDeployer) generateHPA() error {
	hpa := fmt.Sprintf(`apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: %s-hpa
  namespace: %s
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: %s
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 80
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
`,
		kd.config.AppName,
		kd.namespace,
		kd.config.AppName,
	)

	return kd.writeManifest("hpa.yaml", hpa)
}

// generateNamespace 生成 Namespace
func (kd *K8sDeployer) generateNamespace() error {
	namespace := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    name: %s
`,
		kd.namespace,
		kd.namespace,
	)

	return kd.writeManifest("namespace.yaml", namespace)
}

// Apply 应用清单
func (kd *K8sDeployer) Apply(ctx context.Context, manifestPath string) error {
	fmt.Printf("🚀 Applying Kubernetes manifests: %s\n", manifestPath)

	args := []string{"apply", "-f", manifestPath}

	if kd.kubeconfig != "" {
		args = append(args, "--kubeconfig", kd.kubeconfig)
	}

	if kd.dryRun {
		fmt.Printf("[DRY RUN] kubectl %s\n", strings.Join(args, " "))
		return nil
	}

	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Delete 删除资源
func (kd *K8sDeployer) Delete(ctx context.Context, manifestPath string) error {
	fmt.Printf("🗑️  Deleting Kubernetes resources: %s\n", manifestPath)

	args := []string{"delete", "-f", manifestPath}

	if kd.kubeconfig != "" {
		args = append(args, "--kubeconfig", kd.kubeconfig)
	}

	if kd.dryRun {
		fmt.Printf("[DRY RUN] kubectl %s\n", strings.Join(args, " "))
		return nil
	}

	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// GetPods 获取 Pod 列表
func (kd *K8sDeployer) GetPods(ctx context.Context) error {
	fmt.Println("📋 Getting pods...")

	args := []string{"get", "pods", "-n", kd.namespace}

	if kd.kubeconfig != "" {
		args = append(args, "--kubeconfig", kd.kubeconfig)
	}

	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// GetServices 获取服务列表
func (kd *K8sDeployer) GetServices(ctx context.Context) error {
	fmt.Println("📋 Getting services...")

	args := []string{"get", "services", "-n", kd.namespace}

	if kd.kubeconfig != "" {
		args = append(args, "--kubeconfig", kd.kubeconfig)
	}

	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Rollout 滚动更新
func (kd *K8sDeployer) Rollout(ctx context.Context) error {
	fmt.Printf("🔄 Rolling update deployment: %s\n", kd.config.AppName)

	args := []string{"rollout", "restart", "deployment/" + kd.config.AppName, "-n", kd.namespace}

	if kd.kubeconfig != "" {
		args = append(args, "--kubeconfig", kd.kubeconfig)
	}

	if kd.dryRun {
		fmt.Printf("[DRY RUN] kubectl %s\n", strings.Join(args, " "))
		return nil
	}

	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// GetStatus 获取状态
func (kd *K8sDeployer) GetStatus(ctx context.Context) error {
	fmt.Println("📊 Cluster status...")

	// 获取 Pod 状态
	if err := kd.GetPods(ctx); err != nil {
		return err
	}

	// 获取服务状态
	if err := kd.GetServices(ctx); err != nil {
		return err
	}

	// 获取 HPA 状态
	args := []string{"get", "hpa", "-n", kd.namespace}

	if kd.kubeconfig != "" {
		args = append(args, "--kubeconfig", kd.kubeconfig)
	}

	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Scale 扩缩容
func (kd *K8sDeployer) Scale(ctx context.Context, replicas int) error {
	fmt.Printf("Scaling deployment %s to %d replicas\n", kd.config.AppName, replicas)

	args := []string{"scale", "deployment/" + kd.config.AppName, fmt.Sprintf("--replicas=%d", replicas), "-n", kd.namespace}

	if kd.kubeconfig != "" {
		args = append(args, "--kubeconfig", kd.kubeconfig)
	}

	if kd.dryRun {
		fmt.Printf("[DRY RUN] kubectl %s\n", strings.Join(args, " "))
		return nil
	}

	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// writeManifest 写入清单文件
func (kd *K8sDeployer) writeManifest(filename, content string) error {
	path := filepath.Join("k8s", filename)
	return os.WriteFile(path, []byte(content), 0644)
}

// generateEnvVars 生成环境变量
func (kd *K8sDeployer) generateEnvVars() string {
	if len(kd.config.EnvVars) == 0 {
		return ""
	}

	var env strings.Builder
	for k, v := range kd.config.EnvVars {
		env.WriteString(fmt.Sprintf("- name: %s\n", k))
		env.WriteString(fmt.Sprintf("  value: \"%s\"\n", v))
	}

	return env.String()
}

// generateNodeSelector 生成节点选择器
func (kd *K8sDeployer) generateNodeSelector() string {
	if len(kd.config.NodeSelector) == 0 {
		return ""
	}

	var selector strings.Builder
	selector.WriteString("nodeSelector:\n")
	for k, v := range kd.config.NodeSelector {
		selector.WriteString(fmt.Sprintf("  %s: \"%s\"\n", k, v))
	}

	return selector.String()
}

// generateTolerations 生成容忍配置
func (kd *K8sDeployer) generateTolerations() string {
	if len(kd.config.Tolerations) == 0 {
		return ""
	}

	var tolerations strings.Builder
	tolerations.WriteString("tolerations:\n")

	for _, t := range kd.config.Tolerations {
		tolerations.WriteString(fmt.Sprintf("- key: \"%s\"\n", t.Key))
		tolerations.WriteString(fmt.Sprintf("  operator: \"%s\"\n", t.Operator))
		if t.Value != "" {
			tolerations.WriteString(fmt.Sprintf("  value: \"%s\"\n", t.Value))
		}
		tolerations.WriteString(fmt.Sprintf("  effect: \"%s\"\n", t.Effect))
	}

	return tolerations.String()
}

// SetNamespace 设置命名空间
func (kd *K8sDeployer) SetNamespace(namespace string) {
	kd.namespace = namespace
}

// SetDryRun 设置是否为模拟运行
func (kd *K8sDeployer) SetDryRun(dryRun bool) {
	kd.dryRun = dryRun
}

// SetKubeconfig 设置 kubeconfig 路径
func (kd *K8sDeployer) SetKubeconfig(kubeconfig string) {
	kd.kubeconfig = kubeconfig
}
