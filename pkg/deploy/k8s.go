// Package deploy 提供 Kubernetes 部署配置生成。
package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// K8sConfig Kubernetes 配置
type K8sConfig struct {
	Name         string
	Namespace    string
	Image        string
	Replicas     int
	Port         int
	Resources    ResourceConfig
	EnvVars      map[string]string
	HealthCheck  HealthCheckConfig
	Volumes      []VolumeConfig
	ServiceType  string
	IngressHost  string
}

// ResourceConfig 资源配置
type ResourceConfig struct {
	CPURequest    string
	CPULimit      string
	MemoryRequest string
	MemoryLimit   string
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Path            string
	Port            int
	InitialDelay    int
	Period          int
	Timeout         int
	SuccessThreshold int
	FailureThreshold int
}

// VolumeConfig 卷配置
type VolumeConfig struct {
	Name      string
	Path      string
	Size      string
	StorageClass string
}

// DefaultK8sConfig 默认 K8s 配置
func DefaultK8sConfig(name string) *K8sConfig {
	return &K8sConfig{
		Name:       name,
		Namespace:  "default",
		Replicas:   3,
		Port:       8080,
		ServiceType: "ClusterIP",
		Resources: ResourceConfig{
			CPURequest:    "100m",
			CPULimit:      "500m",
			MemoryRequest: "128Mi",
			MemoryLimit:   "512Mi",
		},
		EnvVars: make(map[string]string),
		HealthCheck: HealthCheckConfig{
			Path:            "/health",
			Port:            8080,
			InitialDelay:    10,
			Period:          10,
			Timeout:         5,
			SuccessThreshold: 1,
			FailureThreshold: 3,
		},
		Volumes: make([]VolumeConfig, 0),
	}
}

// K8sGenerator Kubernetes 配置生成器
type K8sGenerator struct {
	config *K8sConfig
}

// NewK8sGenerator 创建 K8s 生成器
func NewK8sGenerator(config *K8sConfig) *K8sGenerator {
	return &K8sGenerator{config: config}
}

// GenerateDeployment 生成 Deployment
func (g *K8sGenerator) GenerateDeployment() string {
	var envVars string
	for k, v := range g.config.EnvVars {
		envVars += fmt.Sprintf(`        - name: %s
          value: "%s"
`, k, v)
	}

	var volumeMounts string
	var volumes string
	for _, vol := range g.config.Volumes {
		volumeMounts += fmt.Sprintf(`        - name: %s
          mountPath: %s
`, vol.Name, vol.Path)
		volumes += fmt.Sprintf(`    - name: %s
      persistentVolumeClaim:
        claimName: %s-pvc
`, vol.Name, vol.Name)
	}

	return fmt.Sprintf(`apiVersion: apps/v1
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
%s
        resources:
          requests:
            cpu: "%s"
            memory: "%s"
          limits:
            cpu: "%s"
            memory: "%s"
        volumeMounts:
%s
        livenessProbe:
          httpGet:
            path: %s
            port: %d
          initialDelaySeconds: %d
          periodSeconds: %d
          timeoutSeconds: %d
          successThreshold: %d
          failureThreshold: %d
        readinessProbe:
          httpGet:
            path: %s
            port: %d
          initialDelaySeconds: %d
          periodSeconds: %d
          timeoutSeconds: %d
          successThreshold: %d
          failureThreshold: %d
      volumes:
%s
`,
		g.config.Name,
		g.config.Namespace,
		g.config.Name,
		g.config.Replicas,
		g.config.Name,
		g.config.Name,
		g.config.Name,
		g.config.Image,
		g.config.Port,
		envVars,
		g.config.Resources.CPURequest,
		g.config.Resources.MemoryRequest,
		g.config.Resources.CPULimit,
		g.config.Resources.MemoryLimit,
		volumeMounts,
		g.config.HealthCheck.Path,
		g.config.HealthCheck.Port,
		g.config.HealthCheck.InitialDelay,
		g.config.HealthCheck.Period,
		g.config.HealthCheck.Timeout,
		g.config.HealthCheck.SuccessThreshold,
		g.config.HealthCheck.FailureThreshold,
		g.config.HealthCheck.Path,
		g.config.HealthCheck.Port,
		g.config.HealthCheck.InitialDelay,
		g.config.HealthCheck.Period,
		g.config.HealthCheck.Timeout,
		g.config.HealthCheck.SuccessThreshold,
		g.config.HealthCheck.FailureThreshold,
		volumes,
	)
}

// GenerateService 生成 Service
func (g *K8sGenerator) GenerateService() string {
	var ports string
	if g.config.Port > 0 {
		ports += fmt.Sprintf(`  - port: %d
    targetPort: %d
    protocol: TCP
    name: http
`, g.config.Port, g.config.Port)
	}

	return fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
  namespace: %s
  labels:
    app: %s
spec:
  type: %s
  ports:
%s
  selector:
    app: %s
`,
		g.config.Name,
		g.config.Namespace,
		g.config.Name,
		g.config.ServiceType,
		ports,
		g.config.Name,
	)
}

// GenerateIngress 生成 Ingress
func (g *K8sGenerator) GenerateIngress() string {
	if g.config.IngressHost == "" {
		return ""
	}

	return fmt.Sprintf(`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: %s
  namespace: %s
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  ingressClassName: nginx
  rules:
  - host: %s
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
		g.config.Name,
		g.config.Namespace,
		g.config.IngressHost,
		g.config.Name,
		g.config.Port,
	)
}

// GenerateConfigMap 生成 ConfigMap
func (g *K8sGenerator) GenerateConfigMap(configName string, data map[string]string) string {
	var configData string
	for k, v := range data {
		// 转义引号
		v = strings.ReplaceAll(v, `"`, `\"`)
		configData += fmt.Sprintf(`  %s: "%s"
`, k, v)
	}

	return fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: %s
  namespace: %s
data:
%s
`,
		configName,
		g.config.Namespace,
		configData,
	)
}

// GeneratePersistentVolumeClaim 生成 PVC
func (g *K8sGenerator) GeneratePersistentVolumeClaim(vol VolumeConfig) string {
	storageClass := ""
	if vol.StorageClass != "" {
		storageClass = fmt.Sprintf(`  storageClassName: %s`, vol.StorageClass)
	}

	return fmt.Sprintf(`apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: %s-pvc
  namespace: %s
spec:
  accessModes:
    - ReadWriteOnce
%s
  resources:
    requests:
      storage: %s
`,
		vol.Name,
		g.config.Namespace,
		storageClass,
		vol.Size,
	)
}

// GenerateNamespace 生成 Namespace
func (g *K8sGenerator) GenerateNamespace() string {
	if g.config.Namespace == "default" {
		return ""
	}

	return fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    name: %s
`,
		g.config.Namespace,
		g.config.Namespace,
	)
}

// GenerateAll 生成所有 K8s 配置
func (g *K8sGenerator) GenerateAll() map[string]string {
	files := make(map[string]string)

	// Namespace
	if ns := g.GenerateNamespace(); ns != "" {
		files["namespace.yaml"] = ns
	}

	// Deployment
	files["deployment.yaml"] = g.GenerateDeployment()

	// Service
	files["service.yaml"] = g.GenerateService()

	// Ingress
	if ingress := g.GenerateIngress(); ingress != "" {
		files["ingress.yaml"] = ingress
	}

	// PVCs
	for _, vol := range g.config.Volumes {
		files[fmt.Sprintf("%s-pvc.yaml", vol.Name)] = g.GeneratePersistentVolumeClaim(vol)
	}

	return files
}

// WriteK8sManifests 写入 K8s 配置文件
func WriteK8sManifests(config *K8sConfig, outputDir string) error {
	generator := NewK8sGenerator(config)
	files := generator.GenerateAll()

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 写入文件
	for filename, content := range files {
		path := filepath.Join(outputDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
	}

	// 生成 kustomization.yml
	kustomization := fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: %s

resources:
`, config.Namespace)

	for filename := range files {
		kustomization += fmt.Sprintf("  - %s\n", filename)
	}

	kustomPath := filepath.Join(outputDir, "kustomization.yaml")
	if err := os.WriteFile(kustomPath, []byte(kustomization), 0644); err != nil {
		return fmt.Errorf("failed to write kustomization.yaml: %w", err)
	}

	return nil
}
