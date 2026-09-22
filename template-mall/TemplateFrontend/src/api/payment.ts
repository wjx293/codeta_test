const PAY_API_BASE = import.meta.env.VITE_PAY_API_BASE_URL || 'http://localhost:8083';

interface PayCallbackResponse {
  code: number;
  message?: string;
}

/** 模拟支付成功：调用 PayWebServer /api/paycallback */
export async function simulatePayment(
  paymentNo: string,
  amount: number,
): Promise<void> {
  const callbackNo = `CB${Date.now()}${Math.random().toString(36).slice(2, 8)}`;

  const res = await fetch(`${PAY_API_BASE}/api/paycallback`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      payment_no: paymentNo,
      amount,
      result: 'success',
      callback_no: callbackNo,
    }),
  });

  const json = (await res.json()) as PayCallbackResponse;
  if (!res.ok || json.code !== 0) {
    throw new Error(json.message || '支付回调失败');
  }
}

export function getPayApiBaseUrl(): string {
  return PAY_API_BASE;
}
