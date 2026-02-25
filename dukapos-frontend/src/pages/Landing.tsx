import { Link } from 'react-router-dom'
import { Button } from '@/components/common'

const features = [
  {
    icon: (
      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 14l6-6m-5.5.5h.01m4.99 5h.01M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16l3.5-2 3.5 2 3.5-2 3.5 2zM10 8.5a.5.5 0 11-1 0 .5.5 0 011 0zm5 5a.5.5 0 11-1 0 .5.5 0 011 0z" />
      </svg>
    ),
    title: 'Point of Sale',
    description: 'Fast and intuitive sales processing with barcode scanning and multiple payment methods'
  },
  {
    icon: (
      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4v10l-8 4m8-4v10l-8 4m8-4V7m-8 4v10m-8-4" />
      </svg>
    ),
    title: 'Inventory Management',
    description: 'Track stock levels, set low-stock alerts, and manage products across multiple categories'
  },
  {
    icon: (
      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
    title: 'M-Pesa Integration',
    description: 'Seamless mobile money payments with automatic reconciliation and transaction tracking'
  },
  {
    icon: (
      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
      </svg>
    ),
    title: 'AI Insights',
    description: 'Smart predictions for stock replenishment based on sales patterns and trends'
  },
  {
    icon: (
      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
      </svg>
    ),
    title: 'Reports & Analytics',
    description: 'Comprehensive reports on sales, profits, and customer behavior with export options'
  },
  {
    icon: (
      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
      </svg>
    ),
    title: 'Staff Management',
    description: 'Manage staff roles, track individual sales performance, and control access permissions'
  }
]

const stats = [
  { value: '10,000+', label: 'Active Shops' },
  { value: 'KSh 500M+', label: 'Monthly Transactions' },
  { value: '99.9%', label: 'Uptime' },
  { value: '4.8/5', label: 'User Rating' }
]

const testimonials = [
  {
    quote: "DukaPOS changed how I run my shop. I track everything from WhatsApp now!",
    name: "Mary Wanjiku",
    role: "Kiosk Owner, Nairobi",
    avatar: "MW"
  },
  {
    quote: "The M-Pesa integration saves me hours every day. No more manual reconciliation.",
    name: "John Ochieng",
    role: "Supermarket Owner, Kisumu",
    avatar: "JO"
  },
  {
    quote: "Low stock alerts have prevented so many lost sales. Best investment this year.",
    grace: "Grace Atieno",
    role: "Wholesale Shop, Mombasa",
    avatar: "GA"
  }
]

export default function Landing() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-surface-50 to-white">
      {/* Header */}
      <header className="fixed top-0 left-0 right-0 bg-white/80 backdrop-blur-xl border-b border-surface-100 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16 lg:h-20">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 lg:w-12 lg:h-12 bg-gradient-to-br from-primary to-emerald-600 rounded-2xl flex items-center justify-center shadow-glow">
                <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
                </svg>
              </div>
              <span className="text-xl lg:text-2xl font-bold bg-gradient-to-r from-primary to-emerald-600 bg-clip-text text-transparent">
                DukaPOS
              </span>
            </div>

            <div className="hidden md:flex items-center gap-2">
              <Link to="/login">
                <Button variant="ghost" size="sm">Sign In</Button>
              </Link>
              <Link to="/register">
                <Button size="sm">Get Started Free</Button>
              </Link>
            </div>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="relative pt-32 pb-20 lg:pt-40 lg:pb-32 overflow-hidden">
        {/* Background decoration */}
        <div className="absolute inset-0 bg-grid-pattern opacity-30" />
        <div className="absolute top-20 left-1/4 w-96 h-96 bg-primary/10 rounded-full blur-3xl" />
        <div className="absolute bottom-20 right-1/4 w-96 h-96 bg-emerald-100 rounded-full blur-3xl" />

        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center max-w-4xl mx-auto">
            <div className="inline-flex items-center gap-2 px-4 py-2 bg-primary/10 text-primary rounded-full text-sm font-medium mb-6 animate-fade-in">
              <span className="w-2 h-2 bg-primary rounded-full animate-pulse" />
              Trusted by 10,000+ Kenyan Businesses
            </div>

            <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold text-surface-900 leading-tight mb-6 animate-slide-up">
              Run Your Shop from{' '}
              <span className="bg-gradient-to-r from-primary to-emerald-500 bg-clip-text text-transparent">
                WhatsApp
              </span>
            </h1>

            <p className="text-lg sm:text-xl text-surface-600 mb-8 max-w-2xl mx-auto animate-slide-up" style={{ animationDelay: '100ms' }}>
              No app to download, no training needed. Manage inventory, track sales, 
              and receive payments - all from the WhatsApp you already use.
            </p>

            <div className="flex flex-col sm:flex-row items-center justify-center gap-4 animate-slide-up" style={{ animationDelay: '200ms' }}>
              <Link to="/register" className="w-full sm:w-auto">
                <Button size="xl" fullWidth className="shadow-glow hover:shadow-xl">
                  Start Free Trial
                  <svg className="w-5 h-5 ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
                  </svg>
                </Button>
              </Link>
              <Link to="/login" className="w-full sm:w-auto">
                <Button variant="outline" size="xl" fullWidth>
                  Watch Demo
                </Button>
              </Link>
            </div>

            <p className="mt-4 text-sm text-surface-500">
              No credit card required • Setup in 2 minutes
            </p>
          </div>

          {/* App Preview Mockup */}
          <div className="mt-16 lg:mt-24 max-w-5xl mx-auto animate-slide-up" style={{ animationDelay: '300ms' }}>
            <div className="relative">
              <div className="absolute -inset-4 bg-gradient-to-r from-primary/20 to-emerald-200/20 rounded-3xl blur-2xl" />
              <div className="relative bg-surface-900 rounded-3xl p-2 lg:p-4 shadow-2xl">
                <div className="bg-white rounded-2xl overflow-hidden">
                  {/* Mock WhatsApp Chat */}
                  <div className="bg-green-50 p-4 border-b">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 bg-primary rounded-full flex items-center justify-center">
                        <svg className="w-6 h-6 text-white" fill="currentColor" viewBox="0 0 24 24">
                          <path d="M17.498 14.382c-.301-.15-1.767-.867-2.04-.966-.273-.101-.473-.15-.673.15-.197.297-.767.965-.94 1.164-.173.199-.347.223-.644.075-.3-.15-1.255-.462-2.39-1.475-.883-.795-1.477-1.77-1.652-2.07-.174-.3-.019-.465.13-.615.136-.135.301-.347.451-.524.146-.181.194-.301.297-.496.1-.21.049-.375.025-.524-.024-.149-.275-1.194.272-2.196l.538-.225c.225-.1.474-.05.644.15l.838.838c.411.41.647.922.723 1.464l.06.449c.06.375-.023.773-.272 1.092-.473.61-1.35 1.365-1.827 1.456-.477.09-1.013.075-1.393.075-.38 0-.748-.015-1.073-.075-.32-.06-1.033-.6-1.966-1.869-.898-1.219-1.496-2.044-1.67-2.387-.176-.348-.059-.537.132-.71l.8-.3c.301-.15.647-.198.987-.075.474.173 1.603.777 1.95 1.214.347.436.374.75.624 1.125.274.374.274.748.274.996z"/>
                        </svg>
                      </div>
                      <div>
                        <p className="font-semibold text-surface-900">DukaPOS</p>
                        <p className="text-xs text-surface-500">Online</p>
                      </div>
                    </div>
                  </div>
                  <div className="p-4 space-y-3 bg-surface-50">
                    <div className="bg-white p-3 rounded-2xl rounded-tl-sm shadow-sm max-w-xs">
                      <p className="text-sm text-surface-700">📊 Daily Report - Feb 22</p>
                      <p className="text-sm font-semibold text-surface-900 mt-1">Sales: KSh 12,450</p>
                      <p className="text-sm text-green-600">Profit: KSh 3,200</p>
                    </div>
                    <div className="bg-green-100 p-3 rounded-2xl rounded-tr-sm max-w-xs ml-auto">
                      <p className="text-sm text-surface-700">add bread 50 30</p>
                    </div>
                    <div className="bg-green-100 p-3 rounded-2xl rounded-tr-sm max-w-xs ml-auto">
                      <p className="text-sm text-surface-700">✅ Added 30 bread @ KSh 50</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Stats Section */}
      <section className="py-16 bg-surface-900">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-8">
            {stats.map((stat, index) => (
              <div key={index} className="text-center">
                <p className="text-3xl lg:text-4xl font-bold text-white mb-2">{stat.value}</p>
                <p className="text-surface-400">{stat.label}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 lg:py-32">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl lg:text-4xl font-bold text-surface-900 mb-4">
              Everything You Need to Run Your Shop
            </h2>
            <p className="text-lg text-surface-600 max-w-2xl mx-auto">
              Powerful features designed specifically for Kenyan small businesses
            </p>
          </div>

          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6 lg:gap-8">
            {features.map((feature, index) => (
              <div 
                key={index}
                className="group p-6 lg:p-8 bg-white rounded-2xl border border-surface-100 hover:border-primary/20 hover:shadow-card-hover transition-all duration-300"
                style={{ animationDelay: `${index * 50}ms` }}
              >
                <div className="w-12 h-12 bg-primary/10 rounded-xl flex items-center justify-center text-primary mb-4 group-hover:bg-primary group-hover:text-white transition-colors">
                  {feature.icon}
                </div>
                <h3 className="text-lg font-semibold text-surface-900 mb-2">{feature.title}</h3>
                <p className="text-surface-600">{feature.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* How It Works */}
      <section className="py-20 lg:py-32 bg-surface-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl lg:text-4xl font-bold text-surface-900 mb-4">
              How It Works
            </h2>
            <p className="text-lg text-surface-600">
              Get started in 3 simple steps
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-8 lg:gap-12">
            <div className="text-center">
              <div className="w-16 h-16 bg-primary text-white rounded-2xl flex items-center justify-center text-2xl font-bold mx-auto mb-4 shadow-glow">
                1
              </div>
              <h3 className="text-lg font-semibold text-surface-900 mb-2">Save Our Number</h3>
              <p className="text-surface-600">Save DukaPOS WhatsApp number and start chatting</p>
            </div>
            <div className="text-center">
              <div className="w-16 h-16 bg-primary text-white rounded-2xl flex items-center justify-center text-2xl font-bold mx-auto mb-4 shadow-glow">
                2
              </div>
              <h3 className="text-lg font-semibold text-surface-900 mb-2">Add Products</h3>
              <p className="text-surface-600">Send commands like "add bread 50 30" to add stock</p>
            </div>
            <div className="text-center">
              <div className="w-16 h-16 bg-primary text-white rounded-2xl flex items-center justify-center text-2xl font-bold mx-auto mb-4 shadow-glow">
                3
              </div>
              <h3 className="text-lg font-semibold text-surface-900 mb-2">Start Selling</h3>
              <p className="text-surface-600">Record sales and get reports instantly</p>
            </div>
          </div>
        </div>
      </section>

      {/* Testimonials */}
      <section className="py-20 lg:py-32">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl lg:text-4xl font-bold text-surface-900 mb-4">
              Loved by Kenyan Shop Owners
            </h2>
          </div>

          <div className="grid md:grid-cols-3 gap-8">
            {testimonials.map((testimonial, index) => (
              <div key={index} className="bg-white p-6 lg:p-8 rounded-2xl border border-surface-100 shadow-card">
                <div className="flex items-center gap-1 mb-4">
                  {[...Array(5)].map((_, i) => (
                    <svg key={i} className="w-5 h-5 text-yellow-400" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                    </svg>
                  ))}
                </div>
                <p className="text-surface-700 mb-6">"{testimonial.quote}"</p>
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 bg-primary/10 rounded-full flex items-center justify-center text-primary font-semibold">
                    {testimonial.avatar}
                  </div>
                  <div>
                    <p className="font-semibold text-surface-900">{testimonial.name}</p>
                    <p className="text-sm text-surface-500">{testimonial.role}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 lg:py-32 bg-gradient-to-br from-primary to-emerald-600 relative overflow-hidden">
        <div className="absolute inset-0 bg-grid-pattern opacity-10" />
        <div className="relative max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h2 className="text-3xl lg:text-4xl font-bold text-white mb-6">
            Ready to Transform Your Business?
          </h2>
          <p className="text-xl text-white/80 mb-8">
            Join thousands of Kenyan shop owners already using DukaPOS
          </p>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <Link to="/register">
              <Button size="xl" className="bg-white text-primary hover:bg-surface-100 shadow-lg">
                Get Started Free
              </Button>
            </Link>
            <Link to="/contact">
              <Button variant="outline" size="xl" className="border-white text-white hover:bg-white/10">
                Talk to Sales
              </Button>
            </Link>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-surface-900 py-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center">
                <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
                </svg>
              </div>
              <span className="text-white font-semibold">DukaPOS</span>
            </div>
            <p className="text-surface-400 text-sm">
              © 2026 DukaPOS. Built with ❤️ in Kenya for Kenyan Businesses 🇰🇪
            </p>
          </div>
        </div>
      </footer>
    </div>
  )
}
