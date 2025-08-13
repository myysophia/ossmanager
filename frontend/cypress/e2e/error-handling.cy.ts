describe('Error Handling and Edge Cases', () => {
  beforeEach(() => {
    cy.loginAsAdmin()
    cy.setupTestBucket('test-bucket')
  })

  describe('Permission Errors', () => {
    it('should handle expired authentication tokens', () => {
      // Mock expired token response
      cy.intercept('GET', '/api/v1/test-bucket?path=/', {
        statusCode: 401,
        body: {
          error: 'Unauthorized',
          message: 'Token has expired',
          code: 'TOKEN_EXPIRED'
        }
      }).as('expiredToken')

      cy.navigateToFileExplorer('test-bucket', '/')
      cy.wait('@expiredToken')

      // Should show token expiration message
      cy.get('[data-testid="auth-error"]').should('be.visible')
      cy.get('[data-testid="error-message"]').should('contain', 'Token has expired')
      
      // Should offer login redirect
      cy.get('[data-testid="login-redirect"]').should('be.visible')
      cy.get('[data-testid="login-redirect"]').click()
      
      // Should redirect to login page
      cy.url().should('include', '/login')
    })

    it('should handle insufficient permissions for bucket access', () => {
      // Mock permission denied response
      cy.intercept('GET', '/api/v1/test-bucket?path=/', {
        statusCode: 403,
        body: {
          error: 'Forbidden',
          message: 'Access denied to bucket',
          code: 'INSUFFICIENT_PERMISSIONS'
        }
      }).as('permissionDenied')

      cy.navigateToFileExplorer('test-bucket', '/')
      cy.wait('@permissionDenied')

      // Should show permission error
      cy.get('[data-testid="permission-error"]').should('be.visible')
      cy.get('[data-testid="error-message"]').should('contain', 'Access denied to bucket')
      
      // Should not show file operations
      cy.get('[data-testid="file-drop-zone"]').should('not.exist')
      cy.get('[data-testid="create-folder"]').should('not.exist')
    })

    it('should handle read-only access restrictions', () => {
      // Mock successful file list but read-only user
      cy.intercept('GET', '/api/v1/test-bucket?path=/', {
        statusCode: 200,
        body: {
          items: [
            {
              name: 'readonly-file.txt',
              path: '/readonly-file.txt',
              isDir: false,
              size: 1024,
              mtime: '2023-12-01T10:00:00Z'
            }
          ],
          total: 1,
          hasMore: false,
          permissions: { read: true, write: false, delete: false }
        }
      }).as('readOnlyAccess')

      // Mock upload attempt with permission error
      cy.intercept('POST', '/api/v1/test-bucket/file**', {
        statusCode: 403,
        body: {
          error: 'Forbidden',
          message: 'Write access denied',
          code: 'WRITE_PERMISSION_DENIED'
        }
      }).as('uploadDenied')

      cy.navigateToFileExplorer('test-bucket', '/')
      cy.wait('@readOnlyAccess')

      // File list should be visible
      cy.verifyFileExists('readonly-file.txt')

      // Upload should be disabled or show error
      cy.uploadFile('test-upload.txt', 'test content')
      cy.wait('@uploadDenied')

      cy.get('[data-testid="permission-error"]').should('be.visible')
      cy.get('[data-testid="error-message"]').should('contain', 'Write access denied')
    })
  })

  describe('Network Failures', () => {
    it('should handle network connection errors', () => {
      cy.navigateToFileExplorer('test-bucket', '/')

      // Simulate network failure
      cy.simulateNetworkFailure()

      // Try to refresh the file list
      cy.get('[data-testid="refresh-files"]').click()

      // Should show network error
      cy.get('[data-testid="network-error"]').should('be.visible')
      cy.get('[data-testid="error-message"]').should('contain', 'Network connection failed')
      
      // Should show retry option
      cy.get('[data-testid="retry-button"]').should('be.visible')

      // Restore network and retry
      cy.restoreNetwork()
      
      cy.intercept('GET', '/api/v1/test-bucket?path=/', {
        statusCode: 200,
        body: { items: [], total: 0, hasMore: false }
      }).as('networkRestored')

      cy.get('[data-testid="retry-button"]').click()
      cy.wait('@networkRestored')

      // Error should disappear
      cy.get('[data-testid="network-error"]').should('not.exist')
      cy.get('[data-testid="file-list"]').should('be.visible')
    })

    it('should handle slow network responses', () => {
      // Mock very slow response
      cy.intercept('GET', '/api/v1/test-bucket?path=/', (req) => {
        req.reply((res) => {
          setTimeout(() => {
            res.send({
              statusCode: 200,
              body: { items: [], total: 0, hasMore: false }
            })
          }, 15000) // 15 second delay
        })
      }).as('slowResponse')

      cy.navigateToFileExplorer('test-bucket', '/')

      // Should show loading indicator
      cy.get('[data-testid="loading-spinner"]').should('be.visible')
      
      // Should show timeout warning after some time
      cy.get('[data-testid="slow-connection-warning"]', { timeout: 10000 }).should('be.visible')
      
      // Should offer cancel option
      cy.get('[data-testid="cancel-request"]').should('be.visible')
      cy.get('[data-testid="cancel-request"]').click()
      
      // Loading should stop
      cy.get('[data-testid="loading-spinner"]').should('not.exist')
    })

    it('should handle intermittent connectivity', () => {
      let requestCount = 0
      
      // Mock intermittent failures
      cy.intercept('GET', '/api/v1/test-bucket?path=/', (req) => {
        requestCount++
        if (requestCount <= 2) {
          req.reply({ statusCode: 500, body: { error: 'Service unavailable' } })
        } else {
          req.reply({
            statusCode: 200,
            body: { items: [], total: 0, hasMore: false }
          })
        }
      }).as('intermittentFailure')

      cy.navigateToFileExplorer('test-bucket', '/')
      cy.wait('@intermittentFailure')

      // Should show error and auto-retry
      cy.get('[data-testid="auto-retry-message"]').should('be.visible')
      cy.get('[data-testid="retry-count"]').should('contain', 'Attempt 1')

      // Wait for automatic retry
      cy.wait('@intermittentFailure')
      cy.get('[data-testid="retry-count"]').should('contain', 'Attempt 2')

      // Eventually should succeed
      cy.wait('@intermittentFailure')
      cy.get('[data-testid="file-list"]').should('be.visible')
      cy.get('[data-testid="auto-retry-message"]').should('not.exist')
    })
  })

  describe('Storage Service Errors', () => {
    it('should handle storage service unavailable', () => {
      cy.intercept('GET', '/api/v1/test-bucket?path=/', {
        statusCode: 503,
        body: {
          error: 'Service Unavailable',
          message: 'Storage service is temporarily unavailable',
          code: 'STORAGE_UNAVAILABLE'
        }
      }).as('storageUnavailable')

      cy.navigateToFileExplorer('test-bucket', '/')
      cy.wait('@storageUnavailable')

      cy.get('[data-testid="service-error"]').should('be.visible')
      cy.get('[data-testid="error-message"]').should('contain', 'Storage service is temporarily unavailable')
      
      // Should show service status page link
      cy.get('[data-testid="status-page-link"]').should('be.visible')
    })

    it('should handle storage quota exceeded', () => {
      cy.intercept('POST', '/api/v1/test-bucket/file**', {
        statusCode: 507,
        body: {
          error: 'Insufficient Storage',
          message: 'Storage quota exceeded',
          code: 'QUOTA_EXCEEDED',
          details: {
            used: '9.8 GB',
            limit: '10 GB',
            remaining: '0.2 GB'
          }
        }
      }).as('quotaExceeded')

      cy.navigateToFileExplorer('test-bucket', '/')
      
      // Try to upload file
      cy.uploadFile('large-file.txt', 'x'.repeat(1000000)) // 1MB file
      cy.wait('@quotaExceeded')

      // Should show quota error with details
      cy.get('[data-testid="quota-error"]').should('be.visible')
      cy.get('[data-testid="quota-details"]').should('contain', '9.8 GB of 10 GB used')
      cy.get('[data-testid="upgrade-storage"]').should('be.visible')
    })

    it('should handle bucket not found errors', () => {
      cy.intercept('GET', '/api/v1/nonexistent-bucket?path=/', {
        statusCode: 404,
        body: {
          error: 'Not Found',
          message: 'Bucket does not exist',
          code: 'BUCKET_NOT_FOUND'
        }
      }).as('bucketNotFound')

      cy.visit('/files/nonexistent-bucket')
      cy.wait('@bucketNotFound')

      cy.get('[data-testid="bucket-error"]').should('be.visible')
      cy.get('[data-testid="error-message"]').should('contain', 'Bucket does not exist')
      
      // Should offer to create bucket or go back
      cy.get('[data-testid="create-bucket-option"]').should('be.visible')
      cy.get('[data-testid="go-back"]').should('be.visible')
    })
  })

  describe('File System Edge Cases', () => {
    it('should handle very long file names', () => {
      const longFileName = 'a'.repeat(255) + '.txt' // 255+ characters
      
      cy.navigateToFileExplorer('test-bucket', '/')
      
      // Try to upload file with very long name
      cy.uploadFile(longFileName, 'test content')
      
      // Should show filename validation error
      cy.get('[data-testid="filename-error"]').should('be.visible')
      cy.get('[data-testid="error-message"]').should('contain', 'Filename too long')
    })

    it('should handle special characters in paths', () => {
      const specialChars = ['<', '>', ':', '"', '|', '?', '*', '\\', '/']
      
      specialChars.forEach(char => {
        const filename = `test${char}file.txt`
        
        cy.navigateToFileExplorer('test-bucket', '/')
        cy.uploadFile(filename, 'test content')
        
        // Should show invalid character error
        cy.get('[data-testid="filename-error"]').should('be.visible')
        cy.get('[data-testid="error-message"]').should('contain', 'Invalid characters in filename')
      })
    })

    it('should handle deeply nested directories', () => {
      // Create very deep path (over 1000 characters)
      const deepPath = Array(50).fill('folder').join('/')
      const fullPath = `/${deepPath}/`
      
      cy.intercept('GET', `/api/v1/test-bucket?path=${encodeURIComponent(fullPath)}`, {
        statusCode: 414,
        body: {
          error: 'URI Too Long',
          message: 'Path exceeds maximum length',
          code: 'PATH_TOO_LONG'
        }
      }).as('pathTooLong')

      cy.visit(`/files/test-bucket?path=${encodeURIComponent(fullPath)}`)
      cy.wait('@pathTooLong')

      cy.get('[data-testid="path-error"]').should('be.visible')
      cy.get('[data-testid="error-message"]').should('contain', 'Path exceeds maximum length')
    })

    it('should handle concurrent modifications', () => {
      cy.intercept('GET', '/api/v1/test-bucket?path=/', {
        statusCode: 200,
        body: {
          items: [
            {
              name: 'shared-file.txt',
              path: '/shared-file.txt',
              isDir: false,
              size: 1024,
              mtime: '2023-12-01T10:00:00Z'
            }
          ],
          total: 1,
          hasMore: false
        }
      }).as('listFiles')

      // Mock delete operation that fails due to concurrent modification
      cy.intercept('DELETE', '/api/v1/test-bucket?path=/shared-file.txt', {
        statusCode: 409,
        body: {
          error: 'Conflict',
          message: 'File was modified by another user',
          code: 'CONCURRENT_MODIFICATION'
        }
      }).as('concurrentModification')

      cy.navigateToFileExplorer('test-bucket', '/')
      cy.wait('@listFiles')

      // Try to delete file
      cy.get('[data-testid="file-item"]')
        .contains('shared-file.txt')
        .rightclick()
      
      cy.get('[data-testid="delete-option"]').click()
      cy.get('[data-testid="confirm-delete"]').click()
      cy.wait('@concurrentModification')

      // Should show conflict error with refresh option
      cy.get('[data-testid="conflict-error"]').should('be.visible')
      cy.get('[data-testid="error-message"]').should('contain', 'File was modified by another user')
      cy.get('[data-testid="refresh-and-retry"]').should('be.visible')
    })
  })

  describe('Memory and Performance Edge Cases', () => {
    it('should handle memory pressure during large operations', () => {
      // Monitor memory usage during test
      let initialMemory: number

      cy.window().then((win) => {
        // @ts-ignore - performance.memory is available in some browsers
        if (win.performance && win.performance.memory) {
          // @ts-ignore
          initialMemory = win.performance.memory.usedJSHeapSize
        }
      })

      cy.navigateToFileExplorer('test-bucket', '/')

      // Simulate loading many files
      const manyFiles = Array.from({ length: 10000 }, (_, i) => ({
        name: `file-${i}.txt`,
        path: `/file-${i}.txt`,
        isDir: false,
        size: Math.floor(Math.random() * 1000000),
        mtime: new Date().toISOString()
      }))

      cy.intercept('GET', '/api/v1/test-bucket?path=/', {
        statusCode: 200,
        body: {
          items: manyFiles,
          total: 10000,
          hasMore: false
        }
      }).as('loadManyFiles')

      cy.get('[data-testid="refresh-files"]').click()
      cy.wait('@loadManyFiles')

      // Should handle large dataset without crashing
      cy.get('[data-testid="file-list"]').should('be.visible')
      cy.get('[data-testid="file-item"]').should('have.length.at.least', 100)

      // Check memory usage hasn't grown excessively
      cy.window().then((win) => {
        // @ts-ignore
        if (win.performance && win.performance.memory && initialMemory) {
          // @ts-ignore
          const currentMemory = win.performance.memory.usedJSHeapSize
          const memoryIncrease = currentMemory - initialMemory
          
          // Memory increase should be reasonable (less than 100MB)
          expect(memoryIncrease).to.be.lessThan(100 * 1024 * 1024)
        }
      })
    })

    it('should handle browser storage limits', () => {
      // Fill up localStorage to test storage limits
      try {
        const bigString = 'x'.repeat(1024 * 1024) // 1MB string
        for (let i = 0; i < 10; i++) {
          localStorage.setItem(`big-item-${i}`, bigString)
        }
      } catch (e) {
        // Expected when storage is full
      }

      cy.navigateToFileExplorer('test-bucket', '/')

      // Try to save user preferences
      cy.get('[data-testid="view-settings"]').click()
      cy.get('[data-testid="grid-view"]').click()

      // Should handle storage quota gracefully
      cy.get('[data-testid="storage-warning"]').should('be.visible')
      cy.get('[data-testid="clear-cache"]').should('be.visible')

      // Clean up
      cy.window().then((win) => {
        win.localStorage.clear()
      })
    })
  })

  describe('Browser Compatibility Edge Cases', () => {
    it('should handle unsupported file API features', () => {
      // Temporarily disable File API
      cy.window().then((win) => {
        // @ts-ignore
        win.File = undefined
        // @ts-ignore
        win.FileReader = undefined
      })

      cy.navigateToFileExplorer('test-bucket', '/')

      // Should show fallback upload interface
      cy.get('[data-testid="legacy-upload"]').should('be.visible')
      cy.get('[data-testid="file-api-unsupported"]').should('be.visible')
    })

    it('should handle JavaScript errors gracefully', () => {
      cy.navigateToFileExplorer('test-bucket', '/')

      // Inject error into page
      cy.window().then((win) => {
        // Cause a JavaScript error
        setTimeout(() => {
          throw new Error('Simulated JavaScript error')
        }, 100)
      })

      // Application should continue working
      cy.get('[data-testid="file-list"]', { timeout: 5000 }).should('be.visible')
      
      // Error boundary should show error message
      cy.get('[data-testid="error-boundary"]').should('be.visible')
      cy.get('[data-testid="recover-button"]').click()
      
      // Should recover from error
      cy.get('[data-testid="file-explorer"]').should('be.visible')
    })
  })
})
