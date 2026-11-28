"use client";

import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { userFacingErrorMessage } from "@/lib/errors";

import { useAddCartItem } from "../hooks";

type AddToCartButtonProps = {
  productId?: string;
  comboId?: string;
  quantity?: number;
  /** In a list beside other items: an outline button, not the page's primary. */
  secondary?: boolean;
};

export function AddToCartButton({
  productId,
  comboId,
  quantity = 1,
  secondary = false,
}: AddToCartButtonProps) {
  const addCartItem = useAddCartItem();

  return (
    <Button
      disabled={addCartItem.isPending}
      size={secondary ? "sm" : "default"}
      type="button"
      variant={secondary ? "outline" : "default"}
      onClick={() => {
        addCartItem.mutate(
          { product_id: productId, combo_id: comboId, quantity },
          { onError: (error) => toast.error(userFacingErrorMessage(error, "We could not add it to your cart. Please try again.")) },
        );
      }}
    >
      {addCartItem.isPending ? "Adding…" : "Add to cart"}
    </Button>
  );
}
