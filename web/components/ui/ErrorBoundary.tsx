'use client'
import { Component, type ReactNode } from 'react'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
  message: string
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, message: '' }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, message: error.message }
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 p-8 text-center">
          <div className="max-w-md">
            <h1 className="text-2xl font-bold text-gray-900 mb-2">
              حدث خطأ غير متوقع / Unexpected Error
            </h1>
            <p className="text-gray-500 text-sm mb-6">{this.state.message}</p>
            <button
              onClick={() => window.location.reload()}
              className="px-4 py-2 bg-brand-600 text-white rounded-lg text-sm hover:bg-brand-700"
            >
              إعادة تحميل / Reload
            </button>
          </div>
        </div>
      )
    }
    return this.props.children
  }
}
