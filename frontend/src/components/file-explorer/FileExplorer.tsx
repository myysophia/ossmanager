'use client';

import React, { useState, useCallback, useEffect } from 'react';
import {
  Box,
  VStack,
  HStack,
  Button,
  IconButton,
  Tooltip,
  useColorModeValue,
  useDisclosure,
  Flex,
  Spacer,
  ButtonGroup,
  useToast,
} from '@chakra-ui/react';
import { FiGrid, FiList, FiPlus, FiSettings } from 'react-icons/fi';

// Import sub-components
import { BreadcrumbNav } from './BreadcrumbNav';
import { ObjectTable } from './ObjectTable';
import { UploadDropzone } from './UploadDropzone';
import { ContextMenu } from './ContextMenu';
import { RenameModal } from './RenameModal';
import { NewFolderModal } from './NewFolderModal';

// Import types
import {
  FileSystemItem,
  BreadcrumbItem,
  UploadProgress,
  SortConfig,
  FileExplorerConfig,
  FileExplorerActions,
} from '../../types/file-explorer';

interface FileExplorerProps {
  /**
   * Current path
   */
  currentPath: string;
  
  /**
   * List of items in current directory
   */
  items: FileSystemItem[];
  
  /**
   * Breadcrumb items for navigation
   */
  breadcrumbs: BreadcrumbItem[];
  
  /**
   * Current upload progress items
   */
  uploads?: UploadProgress[];
  
  /**
   * Configuration options
   */
  config?: Partial<FileExplorerConfig>;
  
  /**
   * Loading states
   */
  isLoading?: boolean;
  isRefreshing?: boolean;
  
  /**
   * View mode
   */
  viewMode?: 'list' | 'grid';
  
  /**
   * Actions and callbacks
   */
  actions: FileExplorerActions;
}

export const FileExplorer: React.FC<FileExplorerProps> = ({
  currentPath,
  items,
  breadcrumbs,
  uploads = [],
  config = {},
  isLoading = false,
  isRefreshing = false,
  viewMode: initialViewMode = 'list',
  actions,
}) => {
  // State management
  const [selectedItems, setSelectedItems] = useState<string[]>([]);
  const [sortConfig, setSortConfig] = useState<SortConfig>({
    key: 'name',
    direction: 'asc',
  });
  const [viewMode, setViewMode] = useState(initialViewMode);
  const [contextMenu, setContextMenu] = useState<{
    isOpen: boolean;
    position: { x: number; y: number };
    items: FileSystemItem[];
  }>({
    isOpen: false,
    position: { x: 0, y: 0 },
    items: [],
  });
  const [showDropzone, setShowDropzone] = useState(true);
  
  // Modal states
  const [renameItem, setRenameItem] = useState<FileSystemItem | null>(null);
  const renameModal = useDisclosure();
  const newFolderModal = useDisclosure();
  
  // Color mode
  const bg = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const toast = useToast();
  
  // Default config
  const finalConfig: FileExplorerConfig = {
    allowMultiSelect: true,
    showHiddenFiles: false,
    allowedFileTypes: undefined,
    maxFileSize: 100 * 1024 * 1024, // 100MB
    maxConcurrentUploads: 5,
    enableDragAndDrop: true,
    enableContextMenu: true,
    ...config,
  };
  
  // Clear selection when path changes
  useEffect(() => {
    setSelectedItems([]);
  }, [currentPath]);
  
  // Handle item selection
  const handleItemSelect = useCallback((itemId: string, isSelected: boolean) => {
    setSelectedItems(prev => {
      if (isSelected) {
        return finalConfig.allowMultiSelect ? [...prev, itemId] : [itemId];
      } else {
        return prev.filter(id => id !== itemId);
      }
    });
  }, [finalConfig.allowMultiSelect]);
  
  // Handle item double click
  const handleItemDoubleClick = useCallback((item: FileSystemItem) => {
    if (item.type === 'folder') {
      actions.onFolderDoubleClick(item);
    } else {
      actions.onFileDoubleClick(item);
    }
  }, [actions]);
  
  // Handle sorting
  const handleSort = useCallback((key: keyof FileSystemItem) => {
    setSortConfig(prev => ({
      key,
      direction: prev.key === key && prev.direction === 'asc' ? 'desc' : 'asc',
    }));
  }, []);
  
  // Handle context menu
  const handleContextMenu = useCallback((e: React.MouseEvent, item?: FileSystemItem) => {
    if (!finalConfig.enableContextMenu) return;
    
    e.preventDefault();
    
    const selectedItemObjects = items.filter(i => selectedItems.includes(i.id));
    const menuItems = item ? [item] : selectedItemObjects;
    
    setContextMenu({
      isOpen: true,
      position: { x: e.clientX, y: e.clientY },
      items: menuItems,
    });
  }, [finalConfig.enableContextMenu, items, selectedItems]);
  
  // Handle rename
  const handleRename = useCallback((item: FileSystemItem) => {
    setRenameItem(item);
    renameModal.onOpen();
  }, [renameModal]);
  
  // Handle rename submit
  const handleRenameSubmit = useCallback(async (item: FileSystemItem, newName: string) => {
    await actions.onRename(item, newName);
    // Refresh the list
    actions.onRefresh();
  }, [actions]);
  
  // Handle delete
  const handleDelete = useCallback((items: FileSystemItem[]) => {
    actions.onDelete(items);
    setSelectedItems([]);
  }, [actions]);
  
  // Handle new folder
  const handleNewFolder = useCallback(() => {
    newFolderModal.onOpen();
  }, [newFolderModal]);
  
  // Handle new folder submit
  const handleNewFolderSubmit = useCallback(async (path: string, name: string) => {
    await actions.onCreateFolder(path, name);
    // Refresh the list
    actions.onRefresh();
  }, [actions]);
  
  // Handle upload
  const handleUpload = useCallback((files: File[]) => {
    actions.onUpload(files, currentPath);
  }, [actions, currentPath]);
  
  // Get selected item objects
  const selectedItemObjects = items.filter(item => selectedItems.includes(item.id));
  
  // Get existing names for validation
  const existingNames = items.map(item => item.name);
  
  return (
    <Box bg={bg} borderRadius="lg" border="1px" borderColor={borderColor} overflow="hidden">
      {/* Header with breadcrumbs and controls */}
      <VStack spacing={0} align="stretch">
        {/* Breadcrumb Navigation */}
        <BreadcrumbNav
          items={breadcrumbs}
          onNavigate={actions.onNavigate}
          onRefresh={actions.onRefresh}
          isRefreshing={isRefreshing}
        />
        
        {/* Toolbar */}
        <Flex
          px={4}
          py={2}
          borderBottom="1px"
          borderColor={borderColor}
          align="center"
          gap={2}
        >
          {/* Left side - Action buttons */}
          <HStack spacing={2}>
            <Button
              size="sm"
              leftIcon={<FiPlus />}
              onClick={handleNewFolder}
              variant="outline"
            >
              新建文件夹
            </Button>
            
            <UploadDropzone
              currentPath={currentPath}
              uploads={uploads}
              onUpload={handleUpload}
              showDropzone={false}
              maxSize={finalConfig.maxFileSize}
              multiple={finalConfig.allowMultiSelect}
            />
            
            {selectedItems.length > 0 && (
              <Button
                size="sm"
                colorScheme="red"
                variant="outline"
                onClick={() => handleDelete(selectedItemObjects)}
              >
                删除 ({selectedItems.length})
              </Button>
            )}
          </HStack>
          
          <Spacer />
          
          {/* Right side - View controls */}
          <HStack spacing={2}>
            <ButtonGroup size="sm" isAttached variant="outline">
              <Tooltip label="列表视图">
                <IconButton
                  aria-label="列表视图"
                  icon={<FiList />}
                  isActive={viewMode === 'list'}
                  onClick={() => setViewMode('list')}
                />
              </Tooltip>
              <Tooltip label="网格视图">
                <IconButton
                  aria-label="网格视图"
                  icon={<FiGrid />}
                  isActive={viewMode === 'grid'}
                  onClick={() => setViewMode('grid')}
                />
              </Tooltip>
            </ButtonGroup>
            
            <Tooltip label="设置">
              <IconButton
                aria-label="设置"
                icon={<FiSettings />}
                size="sm"
                variant="ghost"
              />
            </Tooltip>
            
            <Tooltip label={showDropzone ? '隐藏上传区域' : '显示上传区域'}>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => setShowDropzone(!showDropzone)}
              >
                {showDropzone ? '隐藏上传' : '显示上传'}
              </Button>
            </Tooltip>
          </HStack>
        </Flex>
        
        {/* Upload Dropzone */}
        {showDropzone && (
          <Box p={4} borderBottom="1px" borderColor={borderColor}>
            <UploadDropzone
              currentPath={currentPath}
              uploads={uploads}
              onUpload={handleUpload}
              showDropzone={true}
              maxSize={finalConfig.maxFileSize}
              multiple={finalConfig.allowMultiSelect}
              accept={finalConfig.allowedFileTypes ? 
                Object.fromEntries(finalConfig.allowedFileTypes.map(type => [type, []])) : 
                undefined
              }
            />
          </Box>
        )}
        
        {/* File/Folder Table */}
        <Box
          flex={1}
          onContextMenu={(e) => handleContextMenu(e)}
        >
          <ObjectTable
            items={items}
            selectedItems={selectedItems}
            isLoading={isLoading}
            allowMultiSelect={finalConfig.allowMultiSelect}
            sortConfig={sortConfig}
            onItemSelect={handleItemSelect}
            onItemDoubleClick={handleItemDoubleClick}
            onSort={handleSort}
          />
        </Box>
      </VStack>
      
      {/* Context Menu */}
      <ContextMenu
        isOpen={contextMenu.isOpen}
        selectedItems={contextMenu.items}
        position={contextMenu.position}
        onClose={() => setContextMenu(prev => ({ ...prev, isOpen: false }))}
        onDownload={actions.onDownload}
        onRename={handleRename}
        onDelete={handleDelete}
        onNewFolder={handleNewFolder}
        onRefresh={actions.onRefresh}
      />
      
      {/* Rename Modal */}
      <RenameModal
        isOpen={renameModal.isOpen}
        item={renameItem}
        onClose={() => {
          renameModal.onClose();
          setRenameItem(null);
        }}
        onRename={handleRenameSubmit}
      />
      
      {/* New Folder Modal */}
      <NewFolderModal
        isOpen={newFolderModal.isOpen}
        currentPath={currentPath}
        existingNames={existingNames}
        onClose={newFolderModal.onClose}
        onCreateFolder={handleNewFolderSubmit}
      />
    </Box>
  );
};
