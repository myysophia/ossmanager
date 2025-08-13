'use client';

import React, { useState, useEffect, useRef } from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  Button,
  Input,
  FormControl,
  FormLabel,
  FormErrorMessage,
  useToast,
  HStack,
  Icon,
  Text,
} from '@chakra-ui/react';
import { FiFolder } from 'react-icons/fi';

interface NewFolderModalProps {
  /**
   * Whether modal is open
   */
  isOpen: boolean;
  
  /**
   * Current path where folder will be created
   */
  currentPath: string;
  
  /**
   * Loading state
   */
  isLoading?: boolean;
  
  /**
   * Existing folder names (for validation)
   */
  existingNames?: string[];
  
  /**
   * Callbacks
   */
  onClose: () => void;
  onCreateFolder: (path: string, name: string) => Promise<void>;
}

export const NewFolderModal: React.FC<NewFolderModalProps> = ({
  isOpen,
  currentPath,
  isLoading = false,
  existingNames = [],
  onClose,
  onCreateFolder,
}) => {
  const [folderName, setFolderName] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const toast = useToast();
  
  // Reset form when modal opens/closes
  useEffect(() => {
    if (isOpen) {
      setFolderName('');
      setError('');
      setIsSubmitting(false);
      
      // Focus input
      setTimeout(() => {
        if (inputRef.current) {
          inputRef.current.focus();
        }
      }, 100);
    }
  }, [isOpen]);
  
  // Validate folder name
  const validateName = (name: string): string => {
    if (!name.trim()) {
      return '文件夹名称不能为空';
    }
    
    // Check for invalid characters
    const invalidChars = /[<>:"/\\|?*\x00-\x1f]/g;
    if (invalidChars.test(name)) {
      return '文件夹名称不能包含以下字符: < > : " / \\ | ? *';
    }
    
    // Check for reserved names (Windows)
    const reservedNames = ['CON', 'PRN', 'AUX', 'NUL', 'COM1', 'COM2', 'COM3', 'COM4', 'COM5', 'COM6', 'COM7', 'COM8', 'COM9', 'LPT1', 'LPT2', 'LPT3', 'LPT4', 'LPT5', 'LPT6', 'LPT7', 'LPT8', 'LPT9'];
    if (reservedNames.includes(name.toUpperCase())) {
      return '不能使用保留名称';
    }
    
    // Check for existing names
    if (existingNames.some(existing => existing.toLowerCase() === name.toLowerCase())) {
      return '该名称已存在';
    }
    
    // Check length
    if (name.length > 255) {
      return '文件夹名称长度不能超过 255 个字符';
    }
    
    // Check if starts/ends with dot or space
    if (name.startsWith('.') || name.startsWith(' ') || name.endsWith(' ')) {
      return '文件夹名称不能以点号或空格开头/结尾';
    }
    
    return '';
  };
  
  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const trimmedName = folderName.trim();
    const validationError = validateName(trimmedName);
    
    if (validationError) {
      setError(validationError);
      return;
    }
    
    setIsSubmitting(true);
    setError('');
    
    try {
      await onCreateFolder(currentPath, trimmedName);
      toast({
        title: '创建成功',
        description: `文件夹 "${trimmedName}" 已创建`,
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      onClose();
    } catch (error: any) {
      const errorMessage = error.message || '创建文件夹失败，请重试';
      setError(errorMessage);
      toast({
        title: '创建失败',
        description: errorMessage,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setIsSubmitting(false);
    }
  };
  
  // Handle input change
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFolderName(e.target.value);
    if (error) {
      setError('');
    }
  };
  
  // Handle cancel
  const handleCancel = () => {
    if (!isSubmitting) {
      onClose();
    }
  };
  
  return (
    <Modal
      isOpen={isOpen}
      onClose={handleCancel}
      size="md"
      closeOnOverlayClick={!isSubmitting}
      closeOnEsc={!isSubmitting}
    >
      <ModalOverlay />
      <ModalContent>
        <form onSubmit={handleSubmit}>
          <ModalHeader>
            <HStack spacing={2}>
              <Icon as={FiFolder} color="blue.500" />
              <Text>新建文件夹</Text>
            </HStack>
          </ModalHeader>
          
          <ModalBody>
            <FormControl isInvalid={!!error}>
              <FormLabel>
                文件夹名称
              </FormLabel>
              <Input
                ref={inputRef}
                value={folderName}
                onChange={handleInputChange}
                placeholder="输入文件夹名称"
                isDisabled={isSubmitting || isLoading}
                autoComplete="off"
              />
              <FormErrorMessage>{error}</FormErrorMessage>
            </FormControl>
            
            <Text fontSize="sm" color="gray.500" mt={2}>
              将在当前路径下创建: {currentPath}
            </Text>
          </ModalBody>
          
          <ModalFooter>
            <HStack spacing={3}>
              <Button
                variant="ghost"
                onClick={handleCancel}
                isDisabled={isSubmitting || isLoading}
              >
                取消
              </Button>
              <Button
                type="submit"
                colorScheme="blue"
                isLoading={isSubmitting || isLoading}
                loadingText="创建中..."
                isDisabled={!folderName.trim()}
              >
                创建
              </Button>
            </HStack>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};
