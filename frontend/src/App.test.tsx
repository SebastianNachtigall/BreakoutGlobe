import { render, screen } from '@testing-library/react'
import App from './App'

describe('App', () => {
  it('renders BreakoutGlobe title', () => {
    render(<App />)
    
    const titleElement = screen.getByText(/BreakoutGlobe/i)
    expect(titleElement).toBeInTheDocument()
  })

  it('renders coming soon message', () => {
    render(<App />)
    
    const messageElement = screen.getByText(/Interactive Workshop Platform - Coming Soon/i)
    expect(messageElement).toBeInTheDocument()
  })
})