describe('File Operations - Rename, Delete and Directory Management', () => {
  beforeEach(() => {
    cy.loginAsAdmin()
    cy.setupTestBucket('test-bucket')
    
    // Setup initial file structure
    cy.intercept('GET', '/api/v1/test-bucket?path=/', {
      statusCode: 200,
      body: {
        items: [
          {
            name: 'document.txt',
            path: '/document.txt',
            isDir: false,
            size: 1024,
            mtime: '2023-12-01T10:00:00Z'
          },
          {
            name: 'image.jpg',
            path: '/image.jpg', 
            isDir: false,
            size: 2048576,
            mtime: '2023-12-01T11:00:00Z'
          },
          {
            name: 'folder1',
            path: '/folder1/',
            isDir: true,
            size: 0,
            mtime: '2023-12-01T09:00:00Z'
          }
        ],
        total: 3,
        hasMore: false
      }
    }).as('listInitialFiles')
    
    cy.navigateToFileExplorer('test-bucket', '/')
    cy.wait('@listInitialFiles')
  })

  it('should rename a file successfully', () => {
    const oldFileName = 'document.txt'
    const newFileName = 'renamed-document.txt'

    // Mock rename operation
    cy.intercept('PATCH', '/api/v1/test-bucket/rename', {
      statusCode: 200,
      body: {
        success: true,
        message: `Renamed from ${oldFileName} to ${newFileName} successfully`,
        path: `/${newFileName}`
      }
    }).as('renameFile')

    // Mock updated file list after rename
    cy.intercept('GET', '/api/v1/test-bucket?path=/', {
      statusCode: 200,
      body: {
        items: [
          {
            name: newFileName,
            path: `/${newFileName}`,
            isDir: false,
            size: 1024,
            mtime: new Date().toISOString()
          },
          {
            name: 'image.jpg',
            path: '/image.jpg',
            isDir: false,
            size: 2048576,
            mtime: '2023-12-01T11:00:00Z'
          },
          {
            name: 'folder1',
            path: '/folder1/',
            isDir: true,
            size: 0,
            mtime: '2023-12-01T09:00:00Z'
          }
        ],
        total: 3,
        hasMore: false
      }
    }).as('listAfterRename')

    // Right-click on file to open context menu
    cy.get('[data-testid="file-item"]')
      .contains(oldFileName)
      .rightclick()

    // Select rename option
    cy.get('[data-testid="context-menu"]').should('be.visible')
    cy.get('[data-testid="rename-option"]').click()

    // Rename dialog should appear
    cy.get('[data-testid="rename-dialog"]').should('be.visible')
    cy.get('[data-testid="rename-input"]').should('have.value', oldFileName)

    // Enter new name
    cy.get('[data-testid="rename-input"]').clear().type(newFileName)
    cy.get('[data-testid="confirm-rename"]').click()

    cy.wait('@renameFile')
    cy.wait('@listAfterRename')

    // Verify file has been renamed
    cy.verifyFileExists(newFileName)
    cy.get('[data-testid="file-item"]').should('not.contain', oldFileName)
  })

  it('should handle rename conflicts', () => {
    const existingFileName = 'image.jpg'
    const conflictName = 'image.jpg'

    // Right-click on document.txt to rename it
    cy.get('[data-testid="file-item"]')
      .contains('document.txt')
      .rightclick()

    cy.get('[data-testid="context-menu"]').should('be.visible')
    cy.get('[data-testid="rename-option"]').click()

    // Enter conflicting name
    cy.get('[data-testid="rename-input"]').clear().type(conflictName)
    cy.get('[data-testid="confirm-rename"]').click()

    // Should show conflict error
    cy.get('[data-testid="rename-error"]').should('be.visible')
    cy.get('[data-testid="error-message"]').should('contain', 'File with this name already exists')
  })

  it('should delete a single file', () => {
    const fileToDelete = 'document.txt'

    // Mock delete operation
    cy.intercept('DELETE', `/api/v1/test-bucket?path=/${fileToDelete}`, {
      statusCode: 200,
      body: {
        success: true,
        message: 'File deleted successfully',
        path: `/${fileToDelete}`
      }
    }).as('deleteFile')

    // Mock updated file list after deletion
    cy.intercept('GET', '/api/v1/test-bucket?path=/', {
      statusCode: 200,
      body: {
        items: [
          {
            name: 'image.jpg',
            path: '/image.jpg',
            isDir: false,
            size: 2048576,
            mtime: '2023-12-01T11:00:00Z'
          },
          {
            name: 'folder1',
            path: '/folder1/',
            isDir: true,
            size: 0,
            mtime: '2023-12-01T09:00:00Z'
          }
        ],
        total: 2,
        hasMore: false
      }
    }).as('listAfterDelete')

    // Right-click on file
    cy.get('[data-testid="file-item"]')
      .contains(fileToDelete)
      .rightclick()

    // Select delete option
    cy.get('[data-testid="context-menu"]').should('be.visible')
    cy.get('[data-testid="delete-option"]').click()

    // Confirm deletion dialog
    cy.get('[data-testid="delete-dialog"]').should('be.visible')
    cy.get('[data-testid="delete-message"]').should('contain', fileToDelete)
    cy.get('[data-testid="confirm-delete"]').click()

    cy.wait('@deleteFile')
    cy.wait('@listAfterDelete')

    // Verify file is removed from list
    cy.get('[data-testid="file-item"]').should('not.contain', fileToDelete)
    cy.get('[data-testid="file-item"]').should('have.length', 2)
  })

  it('should delete multiple files', () => {
    const filesToDelete = ['document.txt', 'image.jpg']

    // Mock multiple delete operations
    filesToDelete.forEach(fileName => {
      cy.intercept('DELETE', `/api/v1/test-bucket?path=/${fileName}`, {
        statusCode: 200,
        body: {
          success: true,
          message: 'File deleted successfully',
          path: `/${fileName}`
        }
      }).as(`delete${fileName}`)
    })

    // Mock updated file list after deletions
    cy.intercept('GET', '/api/v1/test-bucket?path=/', {
      statusCode: 200,
      body: {
        items: [
          {
            name: 'folder1',
            path: '/folder1/',
            isDir: true,
            size: 0,
            mtime: '2023-12-01T09:00:00Z'
          }
        ],
        total: 1,
        hasMore: false
      }
    }).as('listAfterMultiDelete')

    // Select multiple files using checkboxes
    filesToDelete.forEach(fileName => {
      cy.get('[data-testid="file-item"]')
        .contains(fileName)
        .within(() => {
          cy.get('[data-testid="file-checkbox"]').check()
        })
    })

    // Verify selection count
    cy.get('[data-testid="selection-count"]').should('contain', '2 items selected')

    // Click bulk delete button
    cy.get('[data-testid="bulk-delete"]').click()

    // Confirm bulk deletion
    cy.get('[data-testid="bulk-delete-dialog"]').should('be.visible')
    cy.get('[data-testid="delete-message"]').should('contain', '2 files')
    cy.get('[data-testid="confirm-bulk-delete"]').click()

    // Wait for all deletions
    filesToDelete.forEach(fileName => {
      cy.wait(`@delete${fileName}`)
    })
    cy.wait('@listAfterMultiDelete')

    // Verify files are removed
    filesToDelete.forEach(fileName => {
      cy.get('[data-testid="file-item"]').should('not.contain', fileName)
    })
    cy.get('[data-testid="file-item"]').should('have.length', 1)
  })

  it('should create a new directory', () => {
    const newDirName = 'new-folder'

    // Mock directory creation
    cy.intercept('POST', '/api/v1/test-bucket/mkdir', {
      statusCode: 200,
      body: {
        success: true,
        message: 'Directory created successfully',
        path: `/${newDirName}/`
      }
    }).as('createDirectory')

    // Mock updated file list with new directory
    cy.intercept('GET', '/api/v1/test-bucket?path=/', {
      statusCode: 200,
      body: {
        items: [
          {
            name: 'document.txt',
            path: '/document.txt',
            isDir: false,
            size: 1024,
            mtime: '2023-12-01T10:00:00Z'
          },
          {
            name: 'image.jpg',
            path: '/image.jpg',
            isDir: false,
            size: 2048576,
            mtime: '2023-12-01T11:00:00Z'
          },
          {
            name: 'folder1',
            path: '/folder1/',
            isDir: true,
            size: 0,
            mtime: '2023-12-01T09:00:00Z'
          },
          {
            name: newDirName,
            path: `/${newDirName}/`,
            isDir: true,
            size: 0,
            mtime: new Date().toISOString()
          }
        ],
        total: 4,
        hasMore: false
      }
    }).as('listAfterDirCreate')

    // Click create folder button
    cy.get('[data-testid="create-folder"]').click()

    // Enter folder name
    cy.get('[data-testid="folder-name-dialog"]').should('be.visible')
    cy.get('[data-testid="folder-name-input"]').type(newDirName)
    cy.get('[data-testid="confirm-create-folder"]').click()

    cy.wait('@createDirectory')
    cy.wait('@listAfterDirCreate')

    // Verify directory appears in list
    cy.get('[data-testid="file-item"]')
      .contains(newDirName)
      .should('be.visible')
      .within(() => {
        cy.get('[data-testid="folder-icon"]').should('be.visible')
      })
  })

  it('should handle delete operation errors', () => {
    const fileToDelete = 'document.txt'

    // Mock failed delete operation
    cy.intercept('DELETE', `/api/v1/test-bucket?path=/${fileToDelete}`, {
      statusCode: 500,
      body: {
        error: 'Internal server error',
        message: 'Failed to delete file: permission denied'
      }
    }).as('failedDelete')

    // Right-click on file
    cy.get('[data-testid="file-item"]')
      .contains(fileToDelete)
      .rightclick()

    cy.get('[data-testid="delete-option"]').click()
    cy.get('[data-testid="confirm-delete"]').click()

    cy.wait('@failedDelete')

    // Verify error message is shown
    cy.get('[data-testid="delete-error"]').should('be.visible')
    cy.get('[data-testid="error-message"]').should('contain', 'permission denied')

    // File should still be in list
    cy.get('[data-testid="file-item"]').should('contain', fileToDelete)
  })

  it('should move files between directories', () => {
    const sourceFile = 'document.txt'
    const targetDir = 'folder1'

    // Setup folder1 contents
    cy.intercept('GET', '/api/v1/test-bucket?path=/folder1/', {
      statusCode: 200,
      body: {
        items: [],
        total: 0,
        hasMore: false
      }
    }).as('listFolder1')

    // Mock move operation (rename to different path)
    cy.intercept('PATCH', '/api/v1/test-bucket/rename', {
      statusCode: 200,
      body: {
        success: true,
        message: `Moved ${sourceFile} to ${targetDir}/`,
        path: `/${targetDir}/${sourceFile}`
      }
    }).as('moveFile')

    // Mock updated root directory after move
    cy.intercept('GET', '/api/v1/test-bucket?path=/', {
      statusCode: 200,
      body: {
        items: [
          {
            name: 'image.jpg',
            path: '/image.jpg',
            isDir: false,
            size: 2048576,
            mtime: '2023-12-01T11:00:00Z'
          },
          {
            name: 'folder1',
            path: '/folder1/',
            isDir: true,
            size: 0,
            mtime: '2023-12-01T09:00:00Z'
          }
        ],
        total: 2,
        hasMore: false
      }
    }).as('listAfterMove')

    // Drag and drop file onto folder
    cy.get('[data-testid="file-item"]')
      .contains(sourceFile)
      .trigger('dragstart')

    cy.get('[data-testid="file-item"]')
      .contains(targetDir)
      .trigger('dragover')
      .trigger('drop')

    // Confirm move dialog
    cy.get('[data-testid="move-dialog"]').should('be.visible')
    cy.get('[data-testid="move-message"]').should('contain', `Move ${sourceFile} to ${targetDir}`)
    cy.get('[data-testid="confirm-move"]').click()

    cy.wait('@moveFile')
    cy.wait('@listAfterMove')

    // Verify file is removed from current directory
    cy.get('[data-testid="file-item"]').should('not.contain', sourceFile)

    // Navigate to target directory to verify file is there
    cy.get('[data-testid="file-item"]')
      .contains(targetDir)
      .dblclick()

    cy.wait('@listFolder1')

    // Mock folder1 with moved file
    cy.intercept('GET', '/api/v1/test-bucket?path=/folder1/', {
      statusCode: 200,
      body: {
        items: [
          {
            name: sourceFile,
            path: `/folder1/${sourceFile}`,
            isDir: false,
            size: 1024,
            mtime: new Date().toISOString()
          }
        ],
        total: 1,
        hasMore: false
      }
    }).as('listFolder1WithFile')

    cy.reload() // Refresh to get updated content
    cy.wait('@listFolder1WithFile')

    cy.verifyFileExists(sourceFile)
  })

  it('should handle file operations with special characters', () => {
    const specialFileName = 'file with spaces & symbols (1).txt'
    const newFileName = 'renamed file with 中文 & émojis 🚀.txt'

    // Add file with special characters to initial list
    cy.intercept('GET', '/api/v1/test-bucket?path=/', {
      statusCode: 200,
      body: {
        items: [
          {
            name: specialFileName,
            path: `/${encodeURIComponent(specialFileName)}`,
            isDir: false,
            size: 1024,
            mtime: '2023-12-01T10:00:00Z'
          }
        ],
        total: 1,
        hasMore: false
      }
    }).as('listSpecialFiles')

    // Mock rename with URL encoding
    cy.intercept('PATCH', '/api/v1/test-bucket/rename', {
      statusCode: 200,
      body: {
        success: true,
        message: `Renamed successfully`,
        path: `/${encodeURIComponent(newFileName)}`
      }
    }).as('renameSpecialFile')

    cy.reload()
    cy.wait('@listSpecialFiles')

    // Test renaming file with special characters
    cy.get('[data-testid="file-item"]')
      .contains(specialFileName)
      .rightclick()

    cy.get('[data-testid="rename-option"]').click()
    cy.get('[data-testid="rename-input"]').clear().type(newFileName)
    cy.get('[data-testid="confirm-rename"]').click()

    cy.wait('@renameSpecialFile')

    // Verify rename was processed correctly
    cy.get('[data-testid="success-message"]').should('be.visible')
  })
})
