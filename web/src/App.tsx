import { Providers } from '@/app/providers'
import { AppRoutes } from '@/app/routes'
import { ErrorBoundary } from '@/components/shared/error-boundary'

function App() {
  return (
    <ErrorBoundary>
      <Providers>
        <AppRoutes />
      </Providers>
    </ErrorBoundary>
  )
}

export default App
