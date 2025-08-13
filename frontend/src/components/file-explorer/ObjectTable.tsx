'use client';

import React from 'react';
import {
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Box,
  Icon,
  Text,
  Checkbox,
  IconButton,
  HStack,
  useColorModeValue,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  MenuDivider,
  Flex,
  Spinner,
  Badge,
  useBreakpointValue,
  Stack,
  Card,
  CardBody,
} from '@chakra-ui/react';
import { ChevronDownIcon, ChevronUpIcon } from '@chakra-ui/icons';
import {
  FiFolder,
  FiFile,
  FiImage,
  FiVideo,
  FiMusic,
  FiArchive,
  FiFileText,
  FiCode,
  FiMoreVertical,
  FiDownload,
  FiEdit2,
  FiTrash2,
  FiShare2,
} from 'react-icons/fi';
import { FileSystemItem, SortConfig, ContextMenuItem } from '../../types/file-explorer';

interface ObjectTableProps {
  /**
   * List of files and folders to display
   */
  items: FileSystemItem[];
  
  /**
   * Selected items
   */
  selectedItems: string[];
  
  /**
   * Loading state
   */
  isLoading?: boolean;
  
  /**
   * Allow multiple selection
   */
  allowMultiSelect?: boolean;
  
  /**
   * Sort configuration
   */
  sortConfig?: SortConfig;
  
  /**
   * Context menu items
   */
  contextMenuItems?: ContextMenuItem[];
  
  /**
   * Callbacks
   */
  onItemSelect?: (itemId: string, isSelected: boolean) => void;
  onItemDoubleClick?: (item: FileSystemItem) => void;
  onSort?: (key: keyof FileSystemItem) => void;
  onContextMenu?: (item: FileSystemItem, menuItems: ContextMenuItem[]) => void;
}

// Helper function to get file icon
const getFileIcon = (item: FileSystemItem) => {
  if (item.type === 'folder') {
    return FiFolder;
  }
  
  const extension = item.name.split('.').pop()?.toLowerCase();
  const mimeType = item.mimeType?.toLowerCase() || '';
  
  if (mimeType.startsWith('image/') || ['jpg', 'jpeg', 'png', 'gif', 'svg', 'webp'].includes(extension || '')) {
    return FiImage;
  }
  
  if (mimeType.startsWith('video/') || ['mp4', 'avi', 'mkv', 'mov', 'wmv'].includes(extension || '')) {
    return FiVideo;
  }
  
  if (mimeType.startsWith('audio/') || ['mp3', 'wav', 'flac', 'aac', 'm4a'].includes(extension || '')) {
    return FiMusic;
  }
  
  if (['zip', 'rar', '7z', 'tar', 'gz'].includes(extension || '')) {
    return FiArchive;
  }
  
  if (['txt', 'md', 'doc', 'docx', 'pdf'].includes(extension || '')) {
    return FiFileText;
  }
  
  if (['js', 'ts', 'jsx', 'tsx', 'html', 'css', 'json', 'xml', 'py', 'java', 'cpp'].includes(extension || '')) {
    return FiCode;
  }
  
  return FiFile;
};

// Helper function to format file size
const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
};

// Helper function to format date
const formatDate = (date: Date): string => {
  const now = new Date();
  const diffTime = now.getTime() - date.getTime();
  const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24));
  
  if (diffDays === 0) {
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
  } else if (diffDays < 7) {
    return `${diffDays}天前`;
  } else {
    return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' });
  }
};

export const ObjectTable: React.FC<ObjectTableProps> = ({
  items,
  selectedItems,
  isLoading = false,
  allowMultiSelect = false,
  sortConfig,
  contextMenuItems = [],
  onItemSelect,
  onItemDoubleClick,
  onSort,
  onContextMenu,
}) => {
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const hoverBg = useColorModeValue('gray.50', 'gray.700');
  const selectedBg = useColorModeValue('blue.50', 'blue.900');
  
  // Responsive view mode
  const isMobile = useBreakpointValue({ base: true, md: false });
  
  // Sort items - folders first, then files
  const sortedItems = React.useMemo(() => {
    const sorted = [...items].sort((a, b) => {
      // Folders first
      if (a.type === 'folder' && b.type === 'file') return -1;
      if (a.type === 'file' && b.type === 'folder') return 1;
      
      // Then sort by configured field
      if (sortConfig) {
        let aVal = a[sortConfig.key];
        let bVal = b[sortConfig.key];
        
        // Handle Date objects
        if (aVal instanceof Date && bVal instanceof Date) {
          aVal = aVal.getTime();
          bVal = bVal.getTime();
        }
        
        if (typeof aVal === 'string' && typeof bVal === 'string') {
          aVal = aVal.toLowerCase();
          bVal = bVal.toLowerCase();
        }
        
        if (aVal < bVal) return sortConfig.direction === 'asc' ? -1 : 1;
        if (aVal > bVal) return sortConfig.direction === 'asc' ? 1 : -1;
      }
      
      return 0;
    });
    
    return sorted;
  }, [items, sortConfig]);
  
  // Handle row click
  const handleRowClick = (item: FileSystemItem, event: React.MouseEvent) => {
    event.preventDefault();
    
    if (allowMultiSelect && onItemSelect) {
      const isSelected = selectedItems.includes(item.id);
      onItemSelect(item.id, !isSelected);
    }
  };
  
  // Handle double click
  const handleDoubleClick = (item: FileSystemItem) => {
    if (onItemDoubleClick) {
      onItemDoubleClick(item);
    }
  };
  
  // Render sort header
  const renderSortHeader = (label: string, key: keyof FileSystemItem) => (
    <Th
      cursor={onSort ? 'pointer' : 'default'}
      onClick={() => onSort && onSort(key)}
      _hover={onSort ? { bg: hoverBg } : undefined}
    >
      <Flex align="center" gap={1}>
        {label}
        {sortConfig?.key === key && (
          <Icon
            as={sortConfig.direction === 'asc' ? ChevronUpIcon : ChevronDownIcon}
            boxSize={3}
          />
        )}
      </Flex>
    </Th>
  );
  
  // Render context menu
  const renderContextMenu = (item: FileSystemItem) => (
    <Menu>
      <MenuButton
        as={IconButton}
        icon={<Icon as={FiMoreVertical} />}
        size="sm"
        variant="ghost"
        aria-label="更多操作"
      />
      <MenuList>
        <MenuItem icon={<Icon as={FiDownload} />}>
          下载
        </MenuItem>
        <MenuItem icon={<Icon as={FiEdit2} />}>
          重命名
        </MenuItem>
        <MenuItem icon={<Icon as={FiShare2} />}>
          分享
        </MenuItem>
        <MenuDivider />
        <MenuItem icon={<Icon as={FiTrash2} />} color="red.500">
          删除
        </MenuItem>
      </MenuList>
    </Menu>
  );
  
  // Loading state
  if (isLoading) {
    return (
      <Flex justify="center" align="center" h="200px">
        <Spinner size="lg" />
      </Flex>
    );
  }
  
  // Empty state
  if (items.length === 0) {
    return (
      <Flex justify="center" align="center" h="200px" direction="column">
        <Icon as={FiFolder} boxSize={12} color="gray.400" mb={4} />
        <Text color="gray.500">此文件夹为空</Text>
      </Flex>
    );
  }
  
  // Mobile card view
  if (isMobile) {
    return (
      <Stack spacing={2} p={4}>
        {sortedItems.map((item) => {
          const isSelected = selectedItems.includes(item.id);
          const IconComponent = getFileIcon(item);
          
          return (
            <Card
              key={item.id}
              cursor="pointer"
              bg={isSelected ? selectedBg : undefined}
              _hover={{ bg: hoverBg }}
              onClick={(e) => handleRowClick(item, e)}
              onDoubleClick={() => handleDoubleClick(item)}
            >
              <CardBody p={3}>
                <Flex align="center" gap={3}>
                  {allowMultiSelect && (
                    <Checkbox
                      isChecked={isSelected}
                      onChange={(e) => onItemSelect && onItemSelect(item.id, e.target.checked)}
                    />
                  )}
                  
                  <Icon
                    as={IconComponent}
                    boxSize={5}
                    color={item.type === 'folder' ? 'blue.500' : 'gray.500'}
                  />
                  
                  <Box flex={1} minW={0}>
                    <Text fontWeight="medium" isTruncated>
                      {item.name}
                    </Text>
                    <HStack spacing={2} fontSize="sm" color="gray.500">
                      {item.type === 'file' && (
                        <Text>{formatFileSize(item.size)}</Text>
                      )}
                      <Text>{formatDate(item.lastModified)}</Text>
                    </HStack>
                  </Box>
                  
                  {renderContextMenu(item)}
                </Flex>
              </CardBody>
            </Card>
          );
        })}
      </Stack>
    );
  }
  
  // Desktop table view
  return (
    <Box overflowX="auto">
      <Table variant="simple" size="sm">
        <Thead>
          <Tr>
            {allowMultiSelect && (
              <Th w="50px">
                <Checkbox
                  isChecked={selectedItems.length === items.length && items.length > 0}
                  isIndeterminate={selectedItems.length > 0 && selectedItems.length < items.length}
                  onChange={(e) => {
                    if (onItemSelect) {
                      items.forEach(item => {
                        onItemSelect(item.id, e.target.checked);
                      });
                    }
                  }}
                />
              </Th>
            )}
            {renderSortHeader('名称', 'name')}
            {renderSortHeader('大小', 'size')}
            {renderSortHeader('修改时间', 'lastModified')}
            <Th w="60px">操作</Th>
          </Tr>
        </Thead>
        <Tbody>
          {sortedItems.map((item) => {
            const isSelected = selectedItems.includes(item.id);
            const IconComponent = getFileIcon(item);
            
            return (
              <Tr
                key={item.id}
                cursor="pointer"
                bg={isSelected ? selectedBg : undefined}
                _hover={{ bg: hoverBg }}
                onClick={(e) => handleRowClick(item, e)}
                onDoubleClick={() => handleDoubleClick(item)}
              >
                {allowMultiSelect && (
                  <Td>
                    <Checkbox
                      isChecked={isSelected}
                      onChange={(e) => onItemSelect && onItemSelect(item.id, e.target.checked)}
                    />
                  </Td>
                )}
                
                <Td>
                  <HStack spacing={3}>
                    <Icon
                      as={IconComponent}
                      boxSize={4}
                      color={item.type === 'folder' ? 'blue.500' : 'gray.500'}
                    />
                    <Text isTruncated maxW="300px" title={item.name}>
                      {item.name}
                    </Text>
                    {item.isLoading && <Spinner size="xs" />}
                  </HStack>
                </Td>
                
                <Td>
                  {item.type === 'file' ? (
                    <Text fontSize="sm" color="gray.600">
                      {formatFileSize(item.size)}
                    </Text>
                  ) : (
                    <Badge colorScheme="blue" variant="subtle" size="sm">
                      文件夹
                    </Badge>
                  )}
                </Td>
                
                <Td>
                  <Text fontSize="sm" color="gray.600">
                    {formatDate(item.lastModified)}
                  </Text>
                </Td>
                
                <Td>
                  {renderContextMenu(item)}
                </Td>
              </Tr>
            );
          })}
        </Tbody>
      </Table>
    </Box>
  );
};
