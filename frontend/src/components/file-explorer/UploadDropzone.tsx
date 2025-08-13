'use client';

import React, { useCallback, useState } from 'react';
import {
  Box,
  Button,
  Text,
  VStack,
  HStack,
  Icon,
  useColorModeValue,
  Progress,
  Alert,
  AlertIcon,
  AlertDescription,
  Flex,
  Badge,
  IconButton,
  Tooltip,
  Stack,
  Card,
  CardBody,
} from '@chakra-ui/react';
import { useDropzone } from 'react-dropzone';
import { FiUpload, FiX, FiFile, FiCheck, FiAlertCircle } from 'react-icons/fi';
import { UploadProgress } from '../../types/file-explorer';

interface UploadDropzoneProps {
  /**
   * Current upload path
   */
  currentPath: string;
  
  /**
   * List of current uploads
   */
  uploads: UploadProgress[];
  
  /**
   * Whether upload is enabled
   */
  disabled?: boolean;
  
  /**
   * Accepted file types
   */
  accept?: Record<string, string[]>;
  
  /**
   * Maximum file size in bytes
   */
  maxSize?: number;
  
  /**
   * Maximum number of files
   */
  maxFiles?: number;
  
  /**
   * Whether to allow multiple files
   */
  multiple?: boolean;
  
  /**
   * Whether to show the dropzone area
   */
  showDropzone?: boolean;
  
  /**
   * Callbacks
   */
  onUpload: (files: File[]) => void;
  onCancelUpload?: (uploadId: string) => void;
  onRetryUpload?: (uploadId: string) => void;
  onClearCompleted?: () => void;
}

export const UploadDropzone: React.FC<UploadDropzoneProps> = ({
  currentPath,
  uploads,
  disabled = false,
  accept,
  maxSize = 100 * 1024 * 1024, // 100MB default
  maxFiles = 10,
  multiple = true,
  showDropzone = true,
  onUpload,
  onCancelUpload,
  onRetryUpload,
  onClearCompleted,
}) => {
  const [dragActive, setDragActive] = useState(false);
  
  // Color mode values
  const borderColor = useColorModeValue('gray.300', 'gray.600');
  const activeBorderColor = useColorModeValue('blue.400', 'blue.300');
  const bgColor = useColorModeValue('gray.50', 'gray.700');
  const activeBgColor = useColorModeValue('blue.50', 'blue.900');
  
  // Handle file drop
  const onDrop = useCallback((acceptedFiles: File[], rejectedFiles: any[]) => {
    setDragActive(false);
    
    if (rejectedFiles.length > 0) {
      // Handle rejected files
      console.warn('Some files were rejected:', rejectedFiles);
    }
    
    if (acceptedFiles.length > 0) {
      onUpload(acceptedFiles);
    }
  }, [onUpload]);
  
  // Configure dropzone
  const {
    getRootProps,
    getInputProps,
    isDragActive,
    isDragAccept,
    isDragReject,
    open,
  } = useDropzone({
    onDrop,
    accept,
    maxSize,
    maxFiles,
    multiple,
    disabled,
    onDragEnter: () => setDragActive(true),
    onDragLeave: () => setDragActive(false),
  });
  
  // Format file size
  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
  };
  
  // Get upload statistics
  const uploadStats = React.useMemo(() => {
    const total = uploads.length;
    const completed = uploads.filter(u => u.status === 'completed').length;
    const failed = uploads.filter(u => u.status === 'failed').length;
    const uploading = uploads.filter(u => u.status === 'uploading').length;
    const cancelled = uploads.filter(u => u.status === 'cancelled').length;
    
    return { total, completed, failed, uploading, cancelled };
  }, [uploads]);
  
  // Render upload progress item
  const renderUploadItem = (upload: UploadProgress) => {
    const getStatusColor = () => {
      switch (upload.status) {
        case 'completed': return 'green';
        case 'failed': return 'red';
        case 'cancelled': return 'gray';
        default: return 'blue';
      }
    };
    
    const getStatusIcon = () => {
      switch (upload.status) {
        case 'completed': return FiCheck;
        case 'failed': return FiAlertCircle;
        default: return FiFile;
      }
    };
    
    return (
      <Card key={upload.id} size="sm" mb={2}>
        <CardBody p={3}>
          <Flex align="center" gap={3}>
            <Icon
              as={getStatusIcon()}
              color={`${getStatusColor()}.500`}
              boxSize={4}
            />
            
            <Box flex={1} minW={0}>
              <Text fontSize="sm" fontWeight="medium" isTruncated>
                {upload.file.name}
              </Text>
              <Text fontSize="xs" color="gray.500">
                {formatFileSize(upload.file.size)}
              </Text>
            </Box>
            
            <VStack spacing={1} align="end" minW="80px">
              <Badge
                colorScheme={getStatusColor()}
                variant="subtle"
                fontSize="xs"
              >
                {upload.status === 'uploading' && `${Math.round(upload.progress)}%`}
                {upload.status === 'completed' && '完成'}
                {upload.status === 'failed' && '失败'}
                {upload.status === 'cancelled' && '取消'}
              </Badge>
              
              {upload.status === 'uploading' && (
                <Progress
                  value={upload.progress}
                  size="sm"
                  colorScheme="blue"
                  width="60px"
                />
              )}
            </VStack>
            
            {/* Action buttons */}
            <HStack spacing={1}>
              {upload.status === 'failed' && onRetryUpload && (
                <Tooltip label="重试">
                  <IconButton
                    aria-label="重试"
                    icon={<Icon as={FiUpload} />}
                    size="xs"
                    variant="ghost"
                    onClick={() => onRetryUpload(upload.id)}
                  />
                </Tooltip>
              )}
              
              {(upload.status === 'uploading' || upload.status === 'failed') && onCancelUpload && (
                <Tooltip label="取消">
                  <IconButton
                    aria-label="取消"
                    icon={<Icon as={FiX} />}
                    size="xs"
                    variant="ghost"
                    colorScheme="red"
                    onClick={() => onCancelUpload(upload.id)}
                  />
                </Tooltip>
              )}
            </HStack>
          </Flex>
          
          {upload.status === 'failed' && upload.error && (
            <Alert status="error" size="sm" mt={2}>
              <AlertIcon boxSize={3} />
              <AlertDescription fontSize="xs">
                {upload.error}
              </AlertDescription>
            </Alert>
          )}
        </CardBody>
      </Card>
    );
  };
  
  return (
    <Box>
      {/* Upload Dropzone */}
      {showDropzone && (
        <Box
          {...getRootProps()}
          border="2px dashed"
          borderColor={
            isDragActive
              ? isDragAccept
                ? activeBorderColor
                : 'red.400'
              : borderColor
          }
          bg={
            isDragActive
              ? isDragAccept
                ? activeBgColor
                : 'red.50'
              : bgColor
          }
          borderRadius="lg"
          p={8}
          textAlign="center"
          cursor={disabled ? 'not-allowed' : 'pointer'}
          transition="all 0.2s"
          _hover={
            !disabled
              ? {
                  borderColor: activeBorderColor,
                  bg: activeBgColor,
                }
              : undefined
          }
        >
          <input {...getInputProps()} />
          
          <VStack spacing={4}>
            <Icon
              as={FiUpload}
              boxSize={12}
              color={
                isDragActive
                  ? isDragAccept
                    ? 'blue.500'
                    : 'red.500'
                  : 'gray.400'
              }
            />
            
            <VStack spacing={1}>
              <Text fontSize="lg" fontWeight="medium">
                {isDragActive
                  ? isDragAccept
                    ? '释放文件以上传'
                    : '不支持的文件类型'
                  : '拖拽文件到这里或点击选择'}
              </Text>
              
              <Text fontSize="sm" color="gray.500">
                支持最大 {formatFileSize(maxSize)} 的文件
                {maxFiles > 1 && `，最多 ${maxFiles} 个文件`}
              </Text>
            </VStack>
            
            <Button
              colorScheme="blue"
              variant="outline"
              leftIcon={<Icon as={FiUpload} />}
              onClick={open}
              isDisabled={disabled}
            >
              选择文件
            </Button>
          </VStack>
        </Box>
      )}
      
      {/* Toolbar Upload Button (when dropzone is hidden) */}
      {!showDropzone && (
        <Button
          leftIcon={<Icon as={FiUpload} />}
          colorScheme="blue"
          onClick={open}
          isDisabled={disabled}
          size="sm"
        >
          上传
        </Button>
      )}
      
      {/* Upload Progress */}
      {uploads.length > 0 && (
        <Box mt={4}>
          <Flex justify="space-between" align="center" mb={3}>
            <HStack spacing={2}>
              <Text fontSize="sm" fontWeight="medium">
                上传进度
              </Text>
              {uploadStats.total > 0 && (
                <Badge variant="outline">
                  {uploadStats.completed}/{uploadStats.total}
                </Badge>
              )}
            </HStack>
            
            {uploadStats.completed > 0 && onClearCompleted && (
              <Button
                size="xs"
                variant="ghost"
                onClick={onClearCompleted}
              >
                清除已完成
              </Button>
            )}
          </Flex>
          
          {/* Upload Statistics */}
          {uploadStats.total > 0 && (
            <HStack spacing={4} mb={3} fontSize="xs" color="gray.600">
              {uploadStats.uploading > 0 && (
                <Text>上传中: {uploadStats.uploading}</Text>
              )}
              {uploadStats.completed > 0 && (
                <Text color="green.600">完成: {uploadStats.completed}</Text>
              )}
              {uploadStats.failed > 0 && (
                <Text color="red.600">失败: {uploadStats.failed}</Text>
              )}
              {uploadStats.cancelled > 0 && (
                <Text color="gray.500">取消: {uploadStats.cancelled}</Text>
              )}
            </HStack>
          )}
          
          {/* Upload Items */}
          <Stack spacing={2} maxH="300px" overflowY="auto">
            {uploads.map(renderUploadItem)}
          </Stack>
        </Box>
      )}
    </Box>
  );
};
