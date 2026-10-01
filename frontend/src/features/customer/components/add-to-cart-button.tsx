"use client";

import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { isApiError } from "@/lib/errors";

import { useAddCartItem } from "../hooks";

type AddToCartButtonProps = {
  productId?: string;
  comboId?: string;
  quantity?: number;
  label?: string;
};

export function AddToCartButton({
  productId,
  comboId,
  quantity = 1,
  label = "Add to cart",
}: AddToCartButtonProps) {
  const addCartItem = useAddCartItem();

  return (
    <Button
      disabled={addCartItem.isPending}
      type="button"
      onClick={() => {
        addCartItem.mutate(
          { product_id: productId, combo_id: comboId, quantity },
          { onError: (error) => toast.error(isApiError(error) ? error.message : "Failed to add to cart") },
        );
      }}
    >
      {addCartItem.isPending ? "Adding…" : label}
    </Button>
  );
}
