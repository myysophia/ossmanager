'use client';

import React, { useCallback, useState } from 'react';
import {
  Box,
  VStack,
  HStack,
  Text,
  Progress,
  Button,
  Badge,
  Alert,
  AlertIcon,
  CloseButton,
  Flex,
  Icon,
  useColorModeValue,
} from '@chakra-ui/react';
import { useDropzone } from 'react-dropzone';
import { FiUpload, FiFile, FiX, FiCheck, FiAlertTriangle } from 'react-icons/fi';
import { useFilesWithSWR } from '../lib/hooks/useFilesWithSWR';
import { ToastUtils } from '../lib/toast';

export interface UploadProgress {
  id: string;
  file: File;
  progress: number;
  status: 'uploading' | 'completed' | 'failed' | 'cancelled';
  error?: string;
}

interface FileUploadZoneProps {
  onUploadComplete?: () => void;
  maxFiles?: number;
  maxFileSize?: number;
  acceptedFileTypes?: string[];
  multiple?: boolean;
  disabled?: boolean;
}

export const FileUploadZone: React.FC<FileUploadZoneProps> = ({
  onUploadComplete,
  maxFiles = 10,
  maxFileSize = 100 * 1024 * 1024, // 100MB
  acceptedFileTypes,
  multiple = true,
  disabled = false,
}) => {
  const [uploadProgress, setUploadProgress] = useState<UploadProgress[]>([]);
  const { uploadFile, refreshFiles } = useFilesWithSWR();

  // Colors
  const borderColor = useColorModeValue('gray.300', 'gray.600');
  const hoverBorderColor = useColorModeValue('blue.400', 'blue.300');
  const bgColor = useColorModeValue('gray.50', 'gray.800');
  const hoverBgColor = useColorModeValue('blue.50', 'blue.900');

  /**
   * Handle file drop or selection
   */
  const onDrop = useCallback(
    async (acceptedFiles: File[], rejectedFiles: any[]) => {
      // Handle rejected files
      if (rejectedFiles.length > 0) {
        rejectedFiles.forEach((rejection) => {
          const errors = rejection.errors.map((e: any) => e.message).join(', ');
          ToastUtils.uploadError(rejection.file.name, errors);
        });
      }

      // Limit number of files
      const filesToUpload = acceptedFiles.slice(0, maxFiles);
      if (acceptedFiles.length > maxFiles) {
        ToastUtils.uploadError('', `只能同时上传 ${maxFiles} 个文件`);
      }

      // Process each file
      for (const file of filesToUpload) {
        const uploadId = `${Date.now()}-${file.name}`;
        
        // Add to progress tracking
        setUploadProgress(prev => [
          ...prev,
          {
            id: uploadId,
            file,
            progress: 0,
            status: 'uploading',
          },
        ]);

        try {
          // Start upload with progress tracking
          const xhr = new XMLHttpRequest();
          
          // Track progress
          xhr.upload.addEventListener('progress', (event) => {
            if (event.lengthComputable) {
              const progress = Math.round((event.loaded / event.total) * 100);
              
              setUploadProgress(prev =>
                prev.map(item =>
                  item.id === uploadId
                    ? { ...item, progress }
                    : item
                )
              );

              // Update toast progress
              ToastUtils.uploadProgress(file.name, progress);
            }
          });

          // Use the hook's upload function
          await uploadFile(file);

          // Mark as completed
          setUploadProgress(prev =>
            prev.map(item =>
              item.id === uploadId
                ? { ...item, progress: 100, status: 'completed' }
                : item
            )
          );

          // Auto-remove successful uploads after delay
          setTimeout(() => {
            setUploadProgress(prev => prev.filter(item => item.id !== uploadId));
          }, 3000);

        } catch (error) {
          // Mark as failed
          setUploadProgress(prev =>
            prev.map(item =>
              item.id === uploadId
                ? {
                    ...item,
                    status: 'failed',
                    error: error instanceof Error ? error.message : '上传失败',
                  }
                : item
            )
          );
        }
      }

      // Callback when all uploads complete
      if (onUploadComplete) {
        onUploadComplete();
      }
    },
    [uploadFile, maxFiles, onUploadComplete]
  );

  /**
   * Cancel upload
   */
  const cancelUpload = useCallback((uploadId: string) => {
    setUploadProgress(prev =>
      prev.map(item =>
        item.id === uploadId
          ? { ...item, status: 'cancelled' }
          : item
      )
    );
  }, []);

  /**
   * Remove upload from list
   */
  const removeUpload = useCallback((uploadId: string) => {
    setUploadProgress(prev => prev.filter(item => item.id !== uploadId));
  }, []);

  /**
   * Retry failed upload
   */
  const retryUpload = useCallback(async (uploadId: string) => {
    const upload = uploadProgress.find(item => item.id === uploadId);
    if (!upload) return;

    // Reset status
    setUploadProgress(prev =>
      prev.map(item =>
        item.id === uploadId
          ? { ...item, status: 'uploading', progress: 0, error: undefined }
          : item
      )
    );

    try {
      await uploadFile(upload.file);
      
      setUploadProgress(prev =>
        prev.map(item =>
          item.id === uploadId
            ? { ...item, progress: 100, status: 'completed' }
            : item
        )
      );

      // Auto-remove after delay
      setTimeout(() => {
        setUploadProgress(prev => prev.filter(item => item.id !== uploadId));
      }, 3000);

    } catch (error) {
      setUploadProgress(prev =>
        prev.map(item =>
          item.id === uploadId
            ? {
                ...item,
                status: 'failed',
                error: error instanceof Error ? error.message : '上传失败',
              }
            : item
        )
      );
    }
  }, [uploadFile, uploadProgress]);

  // Configure dropzone
  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: acceptedFileTypes ? 
      acceptedFileTypes.reduce((acc, type) => {
        acc[type] = [];
        return acc;
      }, {} as Record<string, string[]>) : 
      undefined,
    maxSize: maxFileSize,
    multiple,
    disabled,
  });

  /**
   * Format file size
   */
  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  /**
   * Get status color
   */
  const getStatusColor = (status: UploadProgress['status']) => {
    switch (status) {
      case 'completed':
        return 'green';
      case 'failed':
        return 'red';
      case 'cancelled':
        return 'gray';
      default:
        return 'blue';
    }
  };

  /**
   * Get status icon
   */
  const getStatusIcon = (status: UploadProgress['status']) => {
    switch (status) {
      case 'completed':
        return FiCheck;
      case 'failed':
        return FiAlertTriangle;
      case 'cancelled':
        return FiX;
      default:
        return FiUpload;
    }
  };

  return (
    <VStack spacing={4} w="full">
      {/* Upload Zone */}
      <Box
        {...getRootProps()}
        w="full"
        p={8}
        border="2px dashed"
        borderColor={isDragActive ? hoverBorderColor : borderColor}
        borderRadius="lg"
        bg={isDragActive ? hoverBgColor : bgColor}
        cursor={disabled ? 'not-allowed' : 'pointer'}
        transition="all 0.2s"
        _hover={
          !disabled
            ? {
                borderColor: hoverBorderColor,
                bg: hoverBgColor,
              }
            : {}
        }
        opacity={disabled ? 0.5 : 1}
      >
        <input {...getInputProps()} />
        <VStack spacing={4}>
          <Icon as={FiUpload} boxSize={12} color="gray.400" />
          <VStack spacing={2}>
            <Text fontSize="lg" fontWeight="medium">
              {isDragActive
                ? '松开鼠标上传文件'
                : '拖拽文件到这里或点击选择文件'}
            </Text>
            <Text fontSize="sm" color="gray.500">
              {acceptedFileTypes
                ? `支持格式: ${acceptedFileTypes.join(', ')}`
                : '支持所有格式'}
            </Text>
            <Text fontSize="sm" color="gray.500">
              单个文件最大: {formatFileSize(maxFileSize)} | 最多: {maxFiles} 个文件
            </Text>
          </VStack>
        </VStack>
      </Box>

      {/* Upload Progress */}
      {uploadProgress.length > 0 && (
        <VStack spacing={3} w="full">
          <Text fontSize="md" fontWeight="medium" alignSelf="flex-start">
            上传进度 ({uploadProgress.length})
          </Text>
          
          {uploadProgress.map((upload) => (
            <Box
              key={upload.id}
              w="full"
              p={4}
              bg={bgColor}
              borderRadius="md"
              border="1px solid"
              borderColor={borderColor}
            >
              <Flex justify="space-between" align="center" mb={2}>
                <HStack spacing={3} flex={1} minW={0}>
                  <Icon as={FiFile} color="gray.500" />
                  <VStack align="start" spacing={0} flex={1} minW={0}>
                    <Text fontSize="sm" fontWeight="medium" noOfLines={1}>
                      {upload.file.name}
                    </Text>
                    <HStack spacing={2}>
                      <Text fontSize="xs" color="gray.500">
                        {formatFileSize(upload.file.size)}
                      </Text>
                      <Badge colorScheme={getStatusColor(upload.status)} size="sm">
                        {upload.status === 'uploading' && `${upload.progress}%`}
                        {upload.status === 'completed' && '完成'}
                        {upload.status === 'failed' && '失败'}
                        {upload.status === 'cancelled' && '已取消'}
                      </Badge>
                    </HStack>
                  </VStack>
                </HStack>

                {/* Action buttons */}
                <HStack spacing={2}>
                  {upload.status === 'uploading' && (
                    <Button
                      size="xs"
                      variant="ghost"
                      onClick={() => cancelUpload(upload.id)}
                    >
                      取消
                    </Button>
                  )}
                  {upload.status === 'failed' && (
                    <Button
                      size="xs"
                      colorScheme="blue"
                      variant="ghost"
                      onClick={() => retryUpload(upload.id)}
                    >
                      重试
                    </Button>
                  )}
                  <CloseButton
                    size="sm"
                    onClick={() => removeUpload(upload.id)}
                  />
                </HStack>
              </Flex>

              {/* Progress bar */}
              {upload.status === 'uploading' && (
                <Progress
                  value={upload.progress}
                  colorScheme="blue"
                  size="sm"
                  hasStripe
                  isAnimated
                />
              )}

              {/* Error message */}
              {upload.status === 'failed' && upload.error && (
                <Alert status="error" size="sm" mt={2}>
                  <AlertIcon />
                  <Text fontSize="xs">{upload.error}</Text>
                </Alert>
              )}
            </Box>
          ))}
        </VStack>
      )}
    </VStack>
  );
};
