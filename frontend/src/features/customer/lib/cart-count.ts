/** How many items a cart holds: every line's quantity, added up. */
export function cartItemCount(items: readonly { quantity: number }[]): number {
  return items.reduce((total, item) => total + item.quantity, 0);
}
