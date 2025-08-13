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
import { FiFile, FiFolder } from 'react-icons/fi';
import { FileSystemItem } from '../../types/file-explorer';

interface RenameModalProps {
  /**
   * Whether modal is open
   */
  isOpen: boolean;
  
  /**
   * Item to rename
   */
  item: FileSystemItem | null;
  
  /**
   * Loading state
   */
  isLoading?: boolean;
  
  /**
   * Callbacks
   */
  onClose: () => void;
  onRename: (item: FileSystemItem, newName: string) => Promise<void>;
}

export const RenameModal: React.FC<RenameModalProps> = ({
  isOpen,
  item,
  isLoading = false,
  onClose,
  onRename,
}) => {
  const [newName, setNewName] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const toast = useToast();
  
  // Reset form when modal opens/closes or item changes
  useEffect(() => {
    if (isOpen && item) {
      setNewName(item.name);
      setError('');
      setIsSubmitting(false);
      
      // Focus and select filename (without extension for files)
      setTimeout(() => {
        if (inputRef.current) {
          inputRef.current.focus();
          
          if (item.type === 'file') {
            // Select filename without extension
            const lastDotIndex = item.name.lastIndexOf('.');
            if (lastDotIndex > 0) {
              inputRef.current.setSelectionRange(0, lastDotIndex);
            } else {
              inputRef.current.select();
            }
          } else {
            inputRef.current.select();
          }
        }
      }, 100);
    }
  }, [isOpen, item]);
  
  // Validate filename
  const validateName = (name: string): string => {
    if (!name.trim()) {
      return '名称不能为空';
    }
    
    if (name.trim() === item?.name) {
      return '名称未更改';
    }
    
    // Check for invalid characters
    const invalidChars = /[<>:"/\\|?*\x00-\x1f]/g;
    if (invalidChars.test(name)) {
      return '名称不能包含以下字符: < > : " / \\ | ? *';
    }
    
    // Check for reserved names (Windows)
    const reservedNames = ['CON', 'PRN', 'AUX', 'NUL', 'COM1', 'COM2', 'COM3', 'COM4', 'COM5', 'COM6', 'COM7', 'COM8', 'COM9', 'LPT1', 'LPT2', 'LPT3', 'LPT4', 'LPT5', 'LPT6', 'LPT7', 'LPT8', 'LPT9'];
    const nameWithoutExt = name.split('.')[0].toUpperCase();
    if (reservedNames.includes(nameWithoutExt)) {
      return '不能使用保留名称';
    }
    
    // Check length
    if (name.length > 255) {
      return '名称长度不能超过 255 个字符';
    }
    
    return '';
  };
  
  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!item) return;
    
    const trimmedName = newName.trim();
    const validationError = validateName(trimmedName);
    
    if (validationError) {
      setError(validationError);
      return;
    }
    
    setIsSubmitting(true);
    setError('');
    
    try {
      await onRename(item, trimmedName);
      toast({
        title: '重命名成功',
        description: `"${item.name}" 已重命名为 "${trimmedName}"`,
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      onClose();
    } catch (error: any) {
      const errorMessage = error.message || '重命名失败，请重试';
      setError(errorMessage);
      toast({
        title: '重命名失败',
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
    setNewName(e.target.value);
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
  
  if (!item) {
    return null;
  }
  
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
              <Icon
                as={item.type === 'folder' ? FiFolder : FiFile}
                color={item.type === 'folder' ? 'blue.500' : 'gray.500'}
              />
              <Text>重命名{item.type === 'folder' ? '文件夹' : '文件'}</Text>
            </HStack>
          </ModalHeader>
          
          <ModalBody>
            <FormControl isInvalid={!!error}>
              <FormLabel>
                新名称
              </FormLabel>
              <Input
                ref={inputRef}
                value={newName}
                onChange={handleInputChange}
                placeholder="输入新名称"
                isDisabled={isSubmitting || isLoading}
                autoComplete="off"
              />
              <FormErrorMessage>{error}</FormErrorMessage>
            </FormControl>
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
                loadingText="重命名中..."
                isDisabled={!newName.trim() || newName.trim() === item.name}
              >
                重命名
              </Button>
            </HStack>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};
