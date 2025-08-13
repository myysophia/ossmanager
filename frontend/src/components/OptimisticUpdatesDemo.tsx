'use client';

import React, { useState } from 'react';
import {
  Box,
  VStack,
  HStack,
  Button,
  Text,
  Alert,
  AlertIcon,
  Badge,
  Code,
  Divider,
  useColorModeValue,
  Heading,
} from '@chakra-ui/react';
import { FiUpload, FiTrash2, FiDownload, FiRefreshCw } from 'react-icons/fi';
import { useFilesWithSWR } from '../lib/hooks/useFilesWithSWR';
import { optimisticUpdateManager } from '../lib/optimistic-updates';
import { Toast, ToastUtils } from '../lib/toast';

/**
 * Demo component showcasing optimistic updates functionality
 */
export const OptimisticUpdatesDemo: React.FC = () => {
  const [selectedFileId, setSelectedFileId] = useState<number | null>(null);
  const [isSimulatingError, setIsSimulatingError] = useState(false);

  // Use the optimistic updates hook
  const {
    files,
    total,
    loading,
    error,
    uploadFile,
    deleteFile,
    downloadFile,
    refreshFiles,
  } = useFilesWithSWR();

  // Colors for demo
  const bgColor = useColorModeValue('gray.50', 'gray.900');
  const borderColor = useColorModeValue('gray.200', 'gray.600');

  // Get pending actions for display
  const pendingActions = optimisticUpdateManager.getPendingActions();

  /**
   * Demo: Simulate file upload
   */
  const handleDemoUpload = async () => {
    // Create a fake file for demo
    const fakeFile = new File(['demo content'], 'demo-file.txt', {
      type: 'text/plain',
    });

    try {
      if (isSimulatingError) {
        // Simulate API error for demo
        throw new Error('Simulated upload error for demo');
      }
      
      await uploadFile(fakeFile);
      ToastUtils.uploadSuccess('Demo file uploaded successfully!');
    } catch (error) {
      console.error('Demo upload error:', error);
    }
  };

  /**
   * Demo: Simulate file deletion
   */
  const handleDemoDelete = async () => {
    if (!selectedFileId) {
      Toast.warning({ title: 'Please select a file first' });
      return;
    }

    try {
      if (isSimulatingError) {
        // Simulate API error for demo
        throw new Error('Simulated delete error for demo');
      }

      await deleteFile(selectedFileId);
      setSelectedFileId(null);
    } catch (error) {
      console.error('Demo delete error:', error);
    }
  };

  /**
   * Demo: Simulate file download
   */
  const handleDemoDownload = async () => {
    if (!selectedFileId) {
      Toast.warning({ title: 'Please select a file first' });
      return;
    }

    try {
      if (isSimulatingError) {
        // Simulate API error for demo
        throw new Error('Simulated download error for demo');
      }

      await downloadFile(selectedFileId);
    } catch (error) {
      console.error('Demo download error:', error);
    }
  };

  return (
    <VStack spacing={6} p={6} bg={bgColor} borderRadius="lg" border="1px solid" borderColor={borderColor}>
      <Heading size="md">Optimistic Updates Demo</Heading>
      
      {/* Demo Controls */}
      <VStack spacing={4} align="stretch" w="full">
        <HStack justify="space-between">
          <Text fontWeight="medium">Demo Controls</Text>
          <HStack>
            <Button
              size="sm"
              variant={isSimulatingError ? 'solid' : 'outline'}
              colorScheme={isSimulatingError ? 'red' : 'gray'}
              onClick={() => setIsSimulatingError(!isSimulatingError)}
            >
              {isSimulatingError ? 'Simulating Errors' : 'Normal Mode'}
            </Button>
            <Button
              size="sm"
              leftIcon={<FiRefreshCw />}
              onClick={refreshFiles}
              isLoading={loading}
            >
              Refresh Data
            </Button>
          </HStack>
        </HStack>

        {/* Error Simulation Notice */}
        {isSimulatingError && (
          <Alert status="warning">
            <AlertIcon />
            <Text fontSize="sm">
              Error simulation is ON. Operations will fail to demonstrate rollback behavior.
            </Text>
          </Alert>
        )}

        {/* Pending Actions Display */}
        {pendingActions.length > 0 && (
          <Alert status="info">
            <AlertIcon />
            <VStack align="start" spacing={1}>
              <Text fontWeight="medium">Pending Operations:</Text>
              {pendingActions.map((action, index) => (
                <HStack key={index} spacing={2}>
                  <Badge colorScheme="blue" size="sm">
                    {action.action}
                  </Badge>
                  <Text fontSize="sm">{action.target}</Text>
                </HStack>
              ))}
            </VStack>
          </Alert>
        )}

        <Divider />

        {/* Action Buttons */}
        <VStack spacing={3} align="stretch">
          <Text fontWeight="medium">Actions (with Optimistic Updates)</Text>
          
          <HStack spacing={3} wrap="wrap">
            <Button
              leftIcon={<FiUpload />}
              colorScheme="blue"
              onClick={handleDemoUpload}
              isLoading={optimisticUpdateManager.isPending('upload', 'demo-file.txt')}
            >
              Upload Demo File
            </Button>

            <Button
              leftIcon={<FiDownload />}
              colorScheme="green"
              onClick={handleDemoDownload}
              isDisabled={!selectedFileId}
              isLoading={selectedFileId ? optimisticUpdateManager.isPending('download', selectedFileId) : false}
            >
              Download Selected
            </Button>

            <Button
              leftIcon={<FiTrash2 />}
              colorScheme="red"
              onClick={handleDemoDelete}
              isDisabled={!selectedFileId}
              isLoading={selectedFileId ? optimisticUpdateManager.isPending('delete', selectedFileId) : false}
            >
              Delete Selected
            </Button>
          </HStack>
        </VStack>

        <Divider />

        {/* File List */}
        <VStack align="stretch" spacing={3}>
          <HStack justify="space-between">
            <Text fontWeight="medium">Files ({total})</Text>
            {loading && <Badge colorScheme="blue">Loading...</Badge>}
            {error && <Badge colorScheme="red">Error</Badge>}
          </HStack>

          {files.length === 0 ? (
            <Text color="gray.500" textAlign="center" py={4}>
              No files found. Upload a demo file to see optimistic updates in action!
            </Text>
          ) : (
            <VStack align="stretch" spacing={2} maxH="200px" overflowY="auto">
              {files.slice(0, 10).map((file) => (
                <HStack
                  key={file.id}
                  p={3}
                  bg={selectedFileId === file.id ? 'blue.50' : 'white'}
                  borderRadius="md"
                  border="1px solid"
                  borderColor={selectedFileId === file.id ? 'blue.200' : borderColor}
                  cursor="pointer"
                  onClick={() => setSelectedFileId(file.id)}
                  _hover={{ bg: 'gray.50' }}
                >
                  <VStack align="start" spacing={0} flex={1}>
                    <Text fontSize="sm" fontWeight="medium" noOfLines={1}>
                      {decodeURIComponent(file.original_filename)}
                    </Text>
                    <Text fontSize="xs" color="gray.500">
                      {(file.file_size / 1024).toFixed(2)} KB
                    </Text>
                  </VStack>
                  
                  {selectedFileId === file.id && (
                    <Badge colorScheme="blue" size="sm">Selected</Badge>
                  )}
                </HStack>
              ))}
            </VStack>
          )}
        </VStack>

        <Divider />

        {/* How it Works */}
        <VStack align="stretch" spacing={3}>
          <Text fontWeight="medium">How Optimistic Updates Work</Text>
          <VStack align="start" spacing={2} fontSize="sm">
            <Text>1. <strong>Immediate UI Update:</strong> When you click an action, the UI updates instantly</Text>
            <Text>2. <strong>Background API Call:</strong> The actual API request happens in the background</Text>
            <Text>3. <strong>Success:</strong> If successful, the optimistic update is confirmed</Text>
            <Text>4. <strong>Failure:</strong> If failed, the UI rolls back and shows an error</Text>
          </VStack>
        </VStack>

        {/* Example Code */}
        <VStack align="stretch" spacing={2}>
          <Text fontWeight="medium">Example Usage</Text>
          <Code p={3} borderRadius="md" whiteSpace="pre" fontSize="xs">
{`// Upload with optimistic updates
const { uploadFile } = useFilesWithSWR();

const handleUpload = async (file: File) => {
  try {
    // File appears in UI immediately
    await uploadFile(file);
  } catch (error) {
    // File disappears if upload fails
    console.error('Upload failed:', error);
  }
};`}
          </Code>
        </VStack>
      </VStack>
    </VStack>
  );
};
