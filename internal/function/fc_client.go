package function

import (
	"encoding/json"
	"fmt"
	"github.com/aliyun/fc-go-sdk"
	"github.com/myysophia/ossmanager/internal/config"
	"github.com/myysophia/ossmanager/internal/logger"
	"go.uber.org/zap"
)

// FCClient 是阿里云函数计算服务的客户端
type FCClient struct {
	client       *fc.Client
	serviceName  string
	functionName string
}

// MD5CalculationPayload 是触发MD5计算的函数计算负载
type MD5CalculationPayload struct {
	BucketName string `json:"bucket_name"`
	ObjectKey  string `json:"object_key"`
	FileID     uint   `json:"file_id"`
}

// NewFCClient 创建一个新的函数计算客户端
func NewFCClient(config *config.AliyunOSSConfig, _ interface{}) (*FCClient, error) {
	if !config.FunctionCompute.Enabled {
		logger.Info("函数计算在配置中被禁用")
		return nil, nil
	}

	client, err := fc.NewClient(
		config.FunctionCompute.Endpoint,
		config.FunctionCompute.APIVersion,
		config.FunctionCompute.AccessKeyID,
		config.FunctionCompute.AccessKeySecret,
	)
	if err != nil {
		return nil, fmt.Errorf("创建函数计算客户端失败: %w", err)
	}

	return &FCClient{
		client:       client,
		serviceName:  config.FunctionCompute.ServiceName,
		functionName: config.FunctionCompute.FunctionName,
	}, nil
}

// InvokeMD5Calculation 触发函数计算MD5
func (c *FCClient) InvokeMD5Calculation(bucketName, objectKey string, fileID uint) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("函数计算客户端未初始化或被禁用")
	}

	payload := MD5CalculationPayload{
		BucketName: bucketName,
		ObjectKey:  objectKey,
		FileID:     fileID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化负载失败: %w", err)
	}

	logger.Info("调用函数计算计算MD5",
		zap.String("service", c.serviceName),
		zap.String("function", c.functionName),
		zap.String("bucket", bucketName),
		zap.String("object", objectKey),
		zap.Uint("file_id", fileID))

	// 异步调用函数计算服务
	input := fc.NewInvokeFunctionInput(c.serviceName, c.functionName).
		WithPayload(payloadBytes).
		WithAsyncInvocation()

	_, err = c.client.InvokeFunction(input)
	if err != nil {
		return fmt.Errorf("调用函数计算服务失败: %w", err)
	}

	return nil
}
