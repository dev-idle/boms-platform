"use client";

import Link from "next/link";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { ROUTE } from "@/constants/routes";
import { isApiError } from "@/lib/errors";
import { formatDateTime } from "@/lib/validation/datetime";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useOrder } from "../hooks";

type OrderDetailProps = {
  orderId: string;
};

export function OrderDetail({ orderId }: OrderDetailProps) {
  const isValidId = z.string().uuid().safeParse(orderId).success;
  const orderQuery = useOrder(orderId);

  if (!isValidId) {
    return <p className="text-sm text-zinc-500">Invalid order link.</p>;
  }

  if (orderQuery.isPending) {
    return <p className="text-sm text-zinc-500">Loading order…</p>;
  }

  if (orderQuery.isError) {
    const message =
      isApiError(orderQuery.error) && orderQuery.error.status === 404
        ? "Order not found."
        : "Failed to load order.";
    return <p className="text-sm text-red-600">{message}</p>;
  }

  const order = orderQuery.data;
  if (!order) {
    return <p className="text-sm text-zinc-500">Order not found.</p>;
  }

  return (
    <div className="space-y-6">
      <Link href={ROUTE.orders}>
        <Button type="button" variant="outline">
          Back to orders
        </Button>
      </Link>

      <div className="rounded-lg border border-zinc-200 p-6 dark:border-zinc-800">
        <p className="text-sm text-zinc-500">
          Placed {formatDateTime(order.created_at)} · Status: {order.status}
        </p>
        {order.discount_code_snapshot ? (
          <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
            Discount code: {order.discount_code_snapshot}
          </p>
        ) : null}

        <ul className="mt-6 space-y-3">
          {order.items.map((item) => (
            <li
              key={item.id}
              className="flex items-start justify-between gap-4 text-sm"
            >
              <div>
                <p className="font-medium text-zinc-900 dark:text-zinc-50">
                  {item.quantity}× {item.name}
                </p>
                <p className="text-zinc-500">
                  {formatPriceCents(item.unit_price_cents)} each
                </p>
              </div>
              <p className="font-medium">
                {formatPriceCents(item.line_total_cents)}
              </p>
            </li>
          ))}
        </ul>

        <div className="mt-6 space-y-2 border-t border-zinc-200 pt-4 text-sm dark:border-zinc-800">
          <div className="flex justify-between">
            <span className="text-zinc-600 dark:text-zinc-400">Subtotal</span>
            <span>{formatPriceCents(order.subtotal_cents)}</span>
          </div>
          {order.discount_cents > 0 ? (
            <div className="flex justify-between text-emerald-700 dark:text-emerald-400">
              <span>Discount</span>
              <span>-{formatPriceCents(order.discount_cents)}</span>
            </div>
          ) : null}
          <div className="flex justify-between text-base font-semibold">
            <span>Total</span>
            <span>{formatPriceCents(order.total_cents)}</span>
          </div>
        </div>
      </div>
    </div>
  );
}
