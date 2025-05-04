import { createContext, useContext, useEffect, useState } from 'react'

export type Theme = 'dark' | 'light' | 'system'

type ThemeProviderProps = {
  children: React.ReactNode
  defaultTheme?: Theme
  storageKey?: string
}

type ThemeProviderState = {
  theme: Theme
  resolvedTheme: 'dark' | 'light'
  setTheme: (theme: Theme) => void
}

const ThemeProviderContext = createContext<ThemeProviderState | undefined>(
  undefined
)

export function ThemeProvider({
  children,
  defaultTheme = 'light',
  storageKey = 'packet-sentry-web-console-theme',
}: ThemeProviderProps) {
  const getPreferredTheme = (): Theme => {
    const stored = localStorage.getItem(storageKey) as Theme | null
    if (stored === 'light' || stored === 'dark' || stored === 'system') {
      return stored
    }
    return defaultTheme
  }

  const [theme, setThemeState] = useState<Theme>(getPreferredTheme)

  // Set resolvedTheme based on current theme and system preference
  const [resolvedTheme, setResolvedTheme] = useState<'light' | 'dark'>('light')

  useEffect(() => {
    // Determine system's default theme based on preference
    const systemPrefersDark = window.matchMedia(
      '(prefers-color-scheme: dark)'
    ).matches
    const systemTheme: 'dark' | 'light' = systemPrefersDark ? 'dark' : 'light'

    // If 'theme' is 'system', set resolvedTheme based on system preference
    const initialResolvedTheme =
      theme === 'system' ? systemTheme : theme === 'dark' ? 'dark' : 'light'
    setResolvedTheme(initialResolvedTheme)

    // Apply theme to the HTML root element
    const root = window.document.documentElement
    root.classList.remove('light', 'dark')

    if (initialResolvedTheme === 'dark') {
      root.classList.add('dark')
    } else {
      root.classList.add('light')
    }

    // If theme is set to 'system', listen for system preference changes
    if (theme === 'system') {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      const updateTheme = (e: MediaQueryListEvent) => {
        const newTheme = e.matches ? 'dark' : 'light'
        root.classList.remove('light', 'dark')
        root.classList.add(newTheme)
        setResolvedTheme(newTheme)
      }

      mediaQuery.addEventListener('change', updateTheme)

      return () => mediaQuery.removeEventListener('change', updateTheme)
    }
  }, [theme])

  // Set the theme and store in localStorage
  const setTheme = (newTheme: Theme) => {
    localStorage.setItem(storageKey, newTheme)
    setThemeState(newTheme)
  }

  const value: ThemeProviderState = {
    theme,
    resolvedTheme,
    setTheme,
  }

  return (
    <ThemeProviderContext.Provider value={value}>
      {children}
    </ThemeProviderContext.Provider>
  )
}

export const useTheme = () => {
  const context = useContext(ThemeProviderContext)
  if (!context) throw new Error('useTheme must be used within a ThemeProvider')
  return context
}
