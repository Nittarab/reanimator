import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import App from './App'

describe('App', () => {
  it('renders without crashing', () => {
    render(<App />)
    expect(screen.getByText('AI SRE Platform')).toBeInTheDocument()
  })

  it('renders navigation links', () => {
    render(<App />)
    const nav = screen.getByRole('navigation')
    expect(nav).toHaveTextContent('Incidents')
    expect(nav).toHaveTextContent('Configuration')
  })
})
