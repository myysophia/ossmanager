'use client';

import React from 'react';
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  Box,
  Icon,
  useColorModeValue,
  Flex,
  IconButton,
  Tooltip,
} from '@chakra-ui/react';
import { ChevronRightIcon } from '@chakra-ui/icons';
import { FiHome, FiRefreshCw, FiArrowUp } from 'react-icons/fi';
import { BreadcrumbItem as BreadcrumbItemType } from '../../types/file-explorer';

interface BreadcrumbNavProps {
  /**
   * Current path items for breadcrumb
   */
  items: BreadcrumbItemType[];
  
  /**
   * Callback when navigating to a path
   */
  onNavigate: (path: string) => void;
  
  /**
   * Callback when refreshing current directory
   */
  onRefresh?: () => void;
  
  /**
   * Whether refresh is in progress
   */
  isRefreshing?: boolean;
  
  /**
   * Enable navigation to parent directory
   */
  showUpButton?: boolean;
  
  /**
   * Maximum number of items to show before truncating
   */
  maxItems?: number;
}

export const BreadcrumbNav: React.FC<BreadcrumbNavProps> = ({
  items,
  onNavigate,
  onRefresh,
  isRefreshing = false,
  showUpButton = true,
  maxItems = 5,
}) => {
  const bg = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  
  // Handle navigation to parent directory
  const handleNavigateUp = () => {
    if (items.length > 1) {
      const parentPath = items[items.length - 2].path;
      onNavigate(parentPath);
    }
  };
  
  // Truncate items if too many
  const displayItems = React.useMemo(() => {
    if (items.length <= maxItems) {
      return items;
    }
    
    const truncatedItems = [
      items[0], // Root item
      { name: '...', path: '', isEllipsis: true } as BreadcrumbItemType & { isEllipsis: boolean },
      ...items.slice(-maxItems + 2), // Last few items
    ];
    
    return truncatedItems;
  }, [items, maxItems]);
  
  return (
    <Box
      bg={bg}
      borderBottom="1px"
      borderColor={borderColor}
      px={4}
      py={3}
      position="sticky"
      top={0}
      zIndex={10}
    >
      <Flex align="center" gap={2}>
        {/* Home button */}
        <Tooltip label="回到根目录">
          <IconButton
            aria-label="回到根目录"
            icon={<Icon as={FiHome} />}
            size="sm"
            variant="ghost"
            onClick={() => onNavigate('/')}
          />
        </Tooltip>
        
        {/* Up button */}
        {showUpButton && items.length > 1 && (
          <Tooltip label="上级目录">
            <IconButton
              aria-label="上级目录"
              icon={<Icon as={FiArrowUp} />}
              size="sm"
              variant="ghost"
              onClick={handleNavigateUp}
            />
          </Tooltip>
        )}
        
        {/* Breadcrumb */}
        <Breadcrumb
          spacing="8px"
          separator={<ChevronRightIcon color="gray.400" />}
          flex={1}
        >
          {displayItems.map((item, index) => {
            const isLast = index === displayItems.length - 1;
            const isEllipsis = 'isEllipsis' in item && item.isEllipsis;
            
            return (
              <BreadcrumbItem
                key={`${item.path}-${index}`}
                isCurrentPage={isLast}
              >
                {isEllipsis ? (
                  <Box color="gray.400" fontSize="sm">
                    ...
                  </Box>
                ) : isLast ? (
                  <Box
                    color="gray.600"
                    fontWeight="medium"
                    maxWidth="200px"
                    isTruncated
                  >
                    {item.name || '根目录'}
                  </Box>
                ) : (
                  <BreadcrumbLink
                    onClick={() => onNavigate(item.path)}
                    _hover={{ color: 'blue.500' }}
                    maxWidth="150px"
                    isTruncated
                  >
                    {item.name || '根目录'}
                  </BreadcrumbLink>
                )}
              </BreadcrumbItem>
            );
          })}
        </Breadcrumb>
        
        {/* Refresh button */}
        {onRefresh && (
          <Tooltip label="刷新">
            <IconButton
              aria-label="刷新"
              icon={
                <Icon
                  as={FiRefreshCw}
                  transform={isRefreshing ? 'rotate(360deg)' : 'none'}
                  transition="transform 1s ease-in-out"
                />
              }
              size="sm"
              variant="ghost"
              onClick={onRefresh}
              isDisabled={isRefreshing}
            />
          </Tooltip>
        )}
      </Flex>
    </Box>
  );
};
