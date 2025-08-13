describe('File Explorer - Navigation and Basic Operations', () => {
  beforeEach(() => {
    cy.loginAsAdmin()
    cy.setupTestBucket('test-bucket')
  })

  it('should load file explorer and display empty state', () => {
    cy.navigateToFileExplorer('test-bucket', '/')
    
    cy.get('[data-testid="file-explorer"]').should('be.visible')
    cy.get('[data-testid="breadcrumb"]').should('contain', 'test-bucket')
    cy.get('[data-testid="file-list"]').should('be.visible')
    cy.get('[data-testid="empty-state"]').should('be.visible')
  })

  it('should navigate through directory structure', () => {
    // Mock directory structure
    cy.intercept('GET', '/api/v1/test-bucket?path=/', {
      statusCode: 200,
      body: {
        items: [
          { 
            name: 'documents', 
            path: '/documents/',
            isDir: true,
            size: 0,
            mtime: new Date().toISOString()
          },
          {
            name: 'images',
            path: '/images/',  
            isDir: true,
            size: 0,
            mtime: new Date().toISOString()
          }
        ],
        total: 2,
        hasMore: false
      }
    }).as('listRootFiles')

    cy.intercept('GET', '/api/v1/test-bucket?path=/documents/', {
      statusCode: 200, 
      body: {
        items: [
          {
            name: 'readme.txt',
            path: '/documents/readme.txt',
            isDir: false,
            size: 1024,
            mtime: new Date().toISOString()
          }
        ],
        total: 1,
        hasMore: false
      }
    }).as('listDocuments')

    cy.navigateToFileExplorer('test-bucket', '/')
    cy.wait('@listRootFiles')

    // Verify directories are shown
    cy.get('[data-testid="file-item"]').should('have.length', 2)
    cy.get('[data-testid="file-item"]').first().should('contain', 'documents')
    
    // Click on documents directory
    cy.get('[data-testid="file-item"]').first().click()
    cy.wait('@listDocuments')
    
    // Verify navigation
    cy.url().should('include', '/documents')
    cy.get('[data-testid="breadcrumb"]').should('contain', 'documents')
    cy.get('[data-testid="file-item"]').should('contain', 'readme.txt')
  })

  it('should handle navigation with back button', () => {
    cy.intercept('GET', '/api/v1/test-bucket?path=/', {
      statusCode: 200,
      body: {
        items: [
          { 
            name: 'folder1', 
            path: '/folder1/',
            isDir: true,
            size: 0,
            mtime: new Date().toISOString()
          }
        ],
        total: 1,
        hasMore: false
      }
    }).as('listRoot')

    cy.intercept('GET', '/api/v1/test-bucket?path=/folder1/', {
      statusCode: 200,
      body: {
        items: [],
        total: 0, 
        hasMore: false
      }
    }).as('listFolder')

    cy.navigateToFileExplorer('test-bucket', '/')
    cy.wait('@listRoot')
    
    // Navigate to subfolder
    cy.get('[data-testid="file-item"]').first().click()
    cy.wait('@listFolder')
    
    // Use back button
    cy.get('[data-testid="back-button"]').click()
    cy.wait('@listRoot')
    
    // Verify we're back at root
    cy.url().should('not.include', '/folder1')
    cy.get('[data-testid="breadcrumb"]').should('not.contain', 'folder1')
  })

  it('should display file information correctly', () => {
    const testFiles = [
      {
        name: 'document.pdf',
        path: '/document.pdf',
        isDir: false,
        size: 2048576, // 2MB
        mtime: '2023-12-01T10:30:00Z',
        mimeType: 'application/pdf'
      },
      {
        name: 'image.jpg', 
        path: '/image.jpg',
        isDir: false,
        size: 524288, // 512KB
        mtime: '2023-12-02T15:45:00Z',
        mimeType: 'image/jpeg'
      }
    ]

    cy.intercept('GET', '/api/v1/test-bucket?path=/', {
      statusCode: 200,
      body: {
        items: testFiles,
        total: 2,
        hasMore: false
      }
    }).as('listFiles')

    cy.navigateToFileExplorer('test-bucket', '/')
    cy.wait('@listFiles')

    // Verify file names
    cy.get('[data-testid="file-item"]').should('have.length', 2)
    cy.get('[data-testid="file-item"]').first().should('contain', 'document.pdf')
    cy.get('[data-testid="file-item"]').last().should('contain', 'image.jpg')

    // Verify file sizes are displayed
    cy.get('[data-testid="file-item"]').first().should('contain', '2 MB')
    cy.get('[data-testid="file-item"]').last().should('contain', '512 KB')

    // Verify file types/icons
    cy.get('[data-testid="file-item"]').first().within(() => {
      cy.get('[data-testid="file-icon"]').should('have.attr', 'data-type', 'pdf')
    })
  })

  it('should handle pagination for large directories', () => {
    // Generate many files for pagination test
    const files = Array.from({ length: 150 }, (_, i) => ({
      name: `file${i + 1}.txt`,
      path: `/file${i + 1}.txt`,
      isDir: false,
      size: 1024,
      mtime: new Date().toISOString()
    }))

    // First page
    cy.intercept('GET', '/api/v1/test-bucket?path=/&offset=0&limit=100', {
      statusCode: 200,
      body: {
        items: files.slice(0, 100),
        total: 150,
        offset: 0,
        limit: 100,
        hasMore: true
      }
    }).as('listPage1')

    // Second page  
    cy.intercept('GET', '/api/v1/test-bucket?path=/&offset=100&limit=100', {
      statusCode: 200,
      body: {
        items: files.slice(100, 150),
        total: 150,
        offset: 100,
        limit: 100,
        hasMore: false
      }
    }).as('listPage2')

    cy.navigateToFileExplorer('test-bucket', '/')
    cy.wait('@listPage1')

    // Verify first page loaded
    cy.get('[data-testid="file-item"]').should('have.length', 100)
    cy.get('[data-testid="pagination-info"]').should('contain', '1-100 of 150')

    // Click next page
    cy.get('[data-testid="next-page"]').click()
    cy.wait('@listPage2')

    // Verify second page
    cy.get('[data-testid="file-item"]').should('have.length', 50)
    cy.get('[data-testid="pagination-info"]').should('contain', '101-150 of 150')
  })

  it('should search files within directory', () => {
    const allFiles = [
      { name: 'readme.txt', path: '/readme.txt', isDir: false, size: 1024, mtime: new Date().toISOString() },
      { name: 'config.json', path: '/config.json', isDir: false, size: 512, mtime: new Date().toISOString() },
      { name: 'image.png', path: '/image.png', isDir: false, size: 2048, mtime: new Date().toISOString() }
    ]

    cy.intercept('GET', '/api/v1/test-bucket?path=/', {
      statusCode: 200,
      body: { items: allFiles, total: 3, hasMore: false }
    }).as('listAllFiles')

    cy.intercept('GET', '/api/v1/test-bucket?path=/&search=txt', {
      statusCode: 200, 
      body: { 
        items: allFiles.filter(f => f.name.includes('txt')), 
        total: 1, 
        hasMore: false 
      }
    }).as('searchFiles')

    cy.navigateToFileExplorer('test-bucket', '/')
    cy.wait('@listAllFiles')

    // Verify all files shown initially
    cy.get('[data-testid="file-item"]').should('have.length', 3)

    // Search for txt files
    cy.get('[data-testid="search-input"]').type('txt')
    cy.get('[data-testid="search-button"]').click()
    cy.wait('@searchFiles')

    // Verify search results
    cy.get('[data-testid="file-item"]').should('have.length', 1)
    cy.get('[data-testid="file-item"]').should('contain', 'readme.txt')
    
    // Clear search
    cy.get('[data-testid="clear-search"]').click()
    cy.wait('@listAllFiles')
    cy.get('[data-testid="file-item"]').should('have.length', 3)
  })
})
