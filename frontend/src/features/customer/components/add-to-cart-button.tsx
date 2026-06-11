"use client";

import { Button } from "@/components/ui/button";

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
      showArrow
      type="button"
      onClick={() => {
        addCartItem.mutate({
          product_id: productId,
          combo_id: comboId,
          quantity,
        });
      }}
    >
      {addCartItem.isPending ? "Adding…" : label}
    </Button>
  );
}
