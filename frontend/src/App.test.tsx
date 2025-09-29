import { render, screen } from '@testing-library/react'
import { vi } from 'vitest'
import App from './App'

// Mock MapContainer component
vi.mock('./components/MapContainer', () => ({
  MapContainer: vi.fn(() => <div data-testid="map-container">Map Container</div>)
}))

describe.skip('App', () => {
  it('renders BreakoutGlobe title', () => {
    render(<App />)
    
    const titleElement = screen.getByText(/BreakoutGlobe/i)
    expect(titleElement).toBeInTheDocument()
  })

  it('renders Interactive Workshop Platform subtitle', () => {
    render(<App />)
    
    const messageElement = screen.getByText(/Interactive Workshop Platform/i)
    expect(messageElement).toBeInTheDocument()
  })

  it('renders map container', () => {
    render(<App />)
    
    const mapContainer = screen.getByTestId('map-container')
    expect(mapContainer).toBeInTheDocument()
  })

  it('renders status bar with user count', () => {
    render(<App />)
    
    const statusBar = screen.getByText(/Connected Users: 3/i)
    expect(statusBar).toBeInTheDocument()
  })

  it('renders click instruction', () => {
    render(<App />)
    
    const instruction = screen.getByText(/Click anywhere on the map to move your avatar/i)
    expect(instruction).toBeInTheDocument()
  })
})