import { useRef } from 'react'
import type { Sale, Product } from '@/api/types'

interface ReceiptItem {
  name: string
  quantity: number
  unitPrice: number
  total: number
}

interface ReceiptProps {
  items: ReceiptItem[]
  subtotal: number
  tax?: number
  discount?: number
  total: number
  paymentMethod: string
  mpesaReceipt?: string
  shopName: string
  shopPhone?: string
  saleDate?: Date
  invoiceNumber?: string
}

export function Receipt({
  items,
  subtotal,
  tax = 0,
  discount = 0,
  total,
  paymentMethod,
  mpesaReceipt,
  shopName,
  shopPhone,
  saleDate = new Date(),
  invoiceNumber
}: ReceiptProps) {
  const printRef = useRef<HTMLDivElement>(null)

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-KE', {
      style: 'currency',
      currency: 'KES',
      minimumFractionDigits: 0
    }).format(amount)
  }

  const handlePrint = () => {
    const printContent = printRef.current
    if (!printContent) return

    const printWindow = window.open('', '_blank')
    if (!printWindow) return

    printWindow.document.write(`
      <html>
        <head>
          <title>Receipt - ${invoiceNumber || ''}</title>
          <style>
            * { margin: 0; padding: 0; box-sizing: border-box; }
            body { font-family: 'Courier New', monospace; font-size: 12px; padding: 20px; }
            .receipt { max-width: 300px; margin: 0 auto; }
            .header { text-align: center; margin-bottom: 20px; }
            .shop-name { font-size: 18px; font-weight: bold; }
            .divider { border-bottom: 1px dashed #000; margin: 10px 0; }
            .item { display: flex; justify-content: space-between; margin: 5px 0; }
            .total { font-weight: bold; font-size: 16px; }
            .footer { text-align: center; margin-top: 20px; font-size: 10px; }
            @media print { body { padding: 0; } }
          </style>
        </head>
        <body>
          <div class="receipt">
            ${printContent.innerHTML}
          </div>
        </body>
      </html>
    `)
    printWindow.document.close()
    printWindow.print()
    printWindow.close()
  }

  const formatDate = (date: Date) => {
    return date.toLocaleString('en-KE', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  return (
    <div className="space-y-4">
      {/* Receipt Preview */}
      <div 
        ref={printRef}
        className="bg-white p-6 rounded-lg border border-gray-200 max-w-sm mx-auto"
      >
        <div className="text-center mb-4">
          <h2 className="text-lg font-bold">{shopName}</h2>
          {shopPhone && <p className="text-sm text-gray-600">{shopPhone}</p>}
          <p className="text-xs text-gray-500 mt-1">{formatDate(saleDate)}</p>
          {invoiceNumber && <p className="text-xs text-gray-500">#{invoiceNumber}</p>}
        </div>

        <div className="border-t border-b border-dashed border-gray-300 py-4 mb-4">
          {items.map((item, index) => (
            <div key={index} className="flex justify-between text-sm mb-1">
              <div>
                <span>{item.quantity}x {item.name}</span>
              </div>
              <span>{formatCurrency(item.total)}</span>
            </div>
          ))}
        </div>

        <div className="space-y-2 text-sm">
          <div className="flex justify-between">
            <span>Subtotal</span>
            <span>{formatCurrency(subtotal)}</span>
          </div>
          
          {discount > 0 && (
            <div className="flex justify-between text-green-600">
              <span>Discount</span>
              <span>-{formatCurrency(discount)}</span>
            </div>
          )}
          
          {tax > 0 && (
            <div className="flex justify-between">
              <span>Tax</span>
              <span>{formatCurrency(tax)}</span>
            </div>
          )}
          
          <div className="flex justify-between font-bold text-base border-t border-gray-200 pt-2">
            <span>TOTAL</span>
            <span>{formatCurrency(total)}</span>
          </div>
        </div>

        <div className="mt-4 pt-4 border-t border-gray-200 text-center text-sm">
          <p>Payment: {paymentMethod.toUpperCase()}</p>
          {mpesaReceipt && (
            <p className="text-xs text-gray-500">Ref: {mpesaReceipt}</p>
          )}
        </div>

        <div className="mt-4 text-center text-xs text-gray-500">
          <p>Thank you for your business!</p>
          <p>Powered by DukaPOS</p>
        </div>
      </div>

      {/* Action Buttons */}
      <div className="flex gap-2 justify-center">
        <button
          onClick={handlePrint}
          className="flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-dark transition"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
          </svg>
          Print
        </button>
        <button
          onClick={() => {
            const content = printRef.current?.innerText || ''
            navigator.clipboard.writeText(content)
          }}
          className="flex items-center gap-2 px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
          </svg>
          Copy
        </button>
        <button
          onClick={() => {
            const text = `
${shopName}
${formatDate(saleDate)}
${invoiceNumber ? `#${invoiceNumber}` : ''}
${'─'.repeat(20)}
${items.map(i => `${i.quantity}x ${i.name} - ${formatCurrency(i.total)}`).join('\n')}
${'─'.repeat(20)}
Subtotal: ${formatCurrency(subtotal)}
${discount > 0 ? `Discount: -${formatCurrency(discount)}\n` : ''}
${tax > 0 ? `Tax: ${formatCurrency(tax)}\n` : ''}
TOTAL: ${formatCurrency(total)}
${'─'.repeat(20)}
Payment: ${paymentMethod.toUpperCase()}
${mpesaReceipt ? `Ref: ${mpesaReceipt}` : ''}
Thank you!
            `.trim()
            navigator.clipboard.writeText(text)
          }}
          className="flex items-center gap-2 px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
          Share
        </button>
      </div>
    </div>
  )
}

// Simple receipt data mapper
export function createReceiptData(sale: Sale, product?: Product) {
  return {
    items: [{
      name: product?.name || `Product #${sale.product_id}`,
      quantity: sale.quantity,
      unitPrice: sale.unit_price,
      total: sale.total_amount
    }],
    subtotal: sale.total_amount,
    total: sale.total_amount,
    paymentMethod: sale.payment_method,
    mpesaReceipt: sale.mpesa_receipt,
    saleDate: new Date(sale.created_at)
  }
}
