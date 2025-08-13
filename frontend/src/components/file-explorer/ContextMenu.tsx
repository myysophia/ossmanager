'use client';

import React from 'react';
import {
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  MenuDivider,
  MenuGroup,
  Portal,
  Icon,
  useDisclosure,
  Box,
} from '@chakra-ui/react';
import {
  FiDownload,
  FiEdit2,
  FiTrash2,
  FiShare2,
  FiCopy,
  FiMove,
  FiFolder,
  FiInfo,
  FiEye,
  FiRefreshCw,
} from 'react-icons/fi';
import { FileSystemItem, ContextMenuItem } from '../../types/file-explorer';

interface ContextMenuProps {
  /**
   * Selected items for context menu
   */
  selectedItems: FileSystemItem[];
  
  /**
   * Position of the context menu
   */
  position?: { x: number; y: number };
  
  /**
   * Whether menu is open
   */
  isOpen: boolean;
  
  /**
   * Custom menu items
   */
  customItems?: ContextMenuItem[];
  
  /**
   * Enable/disable specific actions
   */
  enabledActions?: {
    download?: boolean;
    rename?: boolean;
    delete?: boolean;
    share?: boolean;
    copy?: boolean;
    move?: boolean;
    preview?: boolean;
    properties?: boolean;
    newFolder?: boolean;
    refresh?: boolean;
  };
  
  /**
   * Callbacks
   */
  onClose: () => void;
  onDownload?: (items: FileSystemItem[]) => void;
  onRename?: (item: FileSystemItem) => void;
  onDelete?: (items: FileSystemItem[]) => void;
  onShare?: (items: FileSystemItem[]) => void;
  onCopy?: (items: FileSystemItem[]) => void;
  onMove?: (items: FileSystemItem[]) => void;
  onPreview?: (item: FileSystemItem) => void;
  onProperties?: (item: FileSystemItem) => void;
  onNewFolder?: () => void;
  onRefresh?: () => void;
}

export const ContextMenu: React.FC<ContextMenuProps> = ({
  selectedItems,
  position,
  isOpen,
  customItems = [],
  enabledActions = {
    download: true,
    rename: true,
    delete: true,
    share: true,
    copy: true,
    move: true,
    preview: true,
    properties: true,
    newFolder: true,
    refresh: true,
  },
  onClose,
  onDownload,
  onRename,
  onDelete,
  onShare,
  onCopy,
  onMove,
  onPreview,
  onProperties,
  onNewFolder,
  onRefresh,
}) => {
  const hasSelectedItems = selectedItems.length > 0;
  const isSingleItem = selectedItems.length === 1;
  const singleItem = isSingleItem ? selectedItems[0] : null;
  const hasFiles = selectedItems.some(item => item.type === 'file');
  const hasFolders = selectedItems.some(item => item.type === 'folder');
  const isMultipleItems = selectedItems.length > 1;
  
  // Handle menu item clicks
  const handleDownload = () => {
    if (onDownload && hasSelectedItems) {
      onDownload(selectedItems);
    }
    onClose();
  };
  
  const handleRename = () => {
    if (onRename && singleItem) {
      onRename(singleItem);
    }
    onClose();
  };
  
  const handleDelete = () => {
    if (onDelete && hasSelectedItems) {
      onDelete(selectedItems);
    }
    onClose();
  };
  
  const handleShare = () => {
    if (onShare && hasSelectedItems) {
      onShare(selectedItems);
    }
    onClose();
  };
  
  const handleCopy = () => {
    if (onCopy && hasSelectedItems) {
      onCopy(selectedItems);
    }
    onClose();
  };
  
  const handleMove = () => {
    if (onMove && hasSelectedItems) {
      onMove(selectedItems);
    }
    onClose();
  };
  
  const handlePreview = () => {
    if (onPreview && singleItem) {
      onPreview(singleItem);
    }
    onClose();
  };
  
  const handleProperties = () => {
    if (onProperties && singleItem) {
      onProperties(singleItem);
    }
    onClose();
  };
  
  const handleNewFolder = () => {
    if (onNewFolder) {
      onNewFolder();
    }
    onClose();
  };
  
  const handleRefresh = () => {
    if (onRefresh) {
      onRefresh();
    }
    onClose();
  };
  
  const handleCustomItem = (item: ContextMenuItem) => {
    item.action();
    onClose();
  };
  
  if (!isOpen) {
    return null;
  }
  
  return (
    <Portal>
      <Box
        position="fixed"
        top={0}
        left={0}
        right={0}
        bottom={0}
        zIndex={9999}
        onClick={onClose}
      >
        <Menu isOpen={isOpen} onClose={onClose}>
          <MenuButton
            position="absolute"
            left={position?.x || 0}
            top={position?.y || 0}
            w="1px"
            h="1px"
            opacity={0}
            pointerEvents="none"
          />
          <MenuList
            position="absolute"
            left={position?.x || 0}
            top={position?.y || 0}
            zIndex={10000}
            minW="180px"
          >
            {/* File/Folder specific actions */}
            {hasSelectedItems && (
              <MenuGroup title={`${selectedItems.length} 项已选择`}>
                {/* Preview - only for single file */}
                {enabledActions.preview && isSingleItem && singleItem?.type === 'file' && (
                  <MenuItem
                    icon={<Icon as={FiEye} />}
                    onClick={handlePreview}
                  >
                    预览
                  </MenuItem>
                )}
                
                {/* Download */}
                {enabledActions.download && onDownload && (
                  <MenuItem
                    icon={<Icon as={FiDownload} />}
                    onClick={handleDownload}
                  >
                    下载 {isMultipleItems ? `(${selectedItems.length}项)` : ''}
                  </MenuItem>
                )}
                
                {/* Rename - only for single item */}
                {enabledActions.rename && isSingleItem && onRename && (
                  <MenuItem
                    icon={<Icon as={FiEdit2} />}
                    onClick={handleRename}
                  >
                    重命名
                  </MenuItem>
                )}
                
                {/* Copy */}
                {enabledActions.copy && onCopy && (
                  <MenuItem
                    icon={<Icon as={FiCopy} />}
                    onClick={handleCopy}
                  >
                    复制
                  </MenuItem>
                )}
                
                {/* Move */}
                {enabledActions.move && onMove && (
                  <MenuItem
                    icon={<Icon as={FiMove} />}
                    onClick={handleMove}
                  >
                    移动
                  </MenuItem>
                )}
                
                {/* Share */}
                {enabledActions.share && onShare && (
                  <MenuItem
                    icon={<Icon as={FiShare2} />}
                    onClick={handleShare}
                  >
                    分享
                  </MenuItem>
                )}
                
                {/* Properties - only for single item */}
                {enabledActions.properties && isSingleItem && onProperties && (
                  <MenuItem
                    icon={<Icon as={FiInfo} />}
                    onClick={handleProperties}
                  >
                    属性
                  </MenuItem>
                )}
                
                <MenuDivider />
                
                {/* Delete */}
                {enabledActions.delete && onDelete && (
                  <MenuItem
                    icon={<Icon as={FiTrash2} />}
                    color="red.500"
                    onClick={handleDelete}
                  >
                    删除 {isMultipleItems ? `(${selectedItems.length}项)` : ''}
                  </MenuItem>
                )}
              </MenuGroup>
            )}
            
            {/* Folder actions (when no items selected or general actions) */}
            {(!hasSelectedItems || customItems.length > 0) && (
              <>
                {hasSelectedItems && <MenuDivider />}
                
                <MenuGroup title="文件夹操作">
                  {/* New Folder */}
                  {enabledActions.newFolder && onNewFolder && (
                    <MenuItem
                      icon={<Icon as={FiFolder} />}
                      onClick={handleNewFolder}
                    >
                      新建文件夹
                    </MenuItem>
                  )}
                  
                  {/* Refresh */}
                  {enabledActions.refresh && onRefresh && (
                    <MenuItem
                      icon={<Icon as={FiRefreshCw} />}
                      onClick={handleRefresh}
                    >
                      刷新
                    </MenuItem>
                  )}
                </MenuGroup>
              </>
            )}
            
            {/* Custom menu items */}
            {customItems.length > 0 && (
              <>
                <MenuDivider />
                <MenuGroup title="更多操作">
                  {customItems.map((item) => (
                    <MenuItem
                      key={item.id}
                      icon={item.icon ? <Icon as={item.icon} /> : undefined}
                      onClick={() => handleCustomItem(item)}
                      isDisabled={item.disabled}
                      color={item.isDanger ? 'red.500' : undefined}
                    >
                      {item.label}
                    </MenuItem>
                  ))}
                </MenuGroup>
              </>
            )}
          </MenuList>
        </Menu>
      </Box>
    </Portal>
  );
};
