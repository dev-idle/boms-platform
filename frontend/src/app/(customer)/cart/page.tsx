import { CartView } from "@/features/customer";

export default function CartPage() {
  return (
    <div>
      <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-50">Cart</h1>
      <p className="mt-2 max-w-prose text-sm text-zinc-600 dark:text-zinc-400">
        Prices and discounts are calculated on the server when your cart loads or
        at checkout.
      </p>
      <div className="mt-6">
        <CartView />
      </div>
    </div>
  );
}
