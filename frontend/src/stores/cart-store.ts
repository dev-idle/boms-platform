import { create } from "zustand";

type CartLine = { productId: string; quantity: number };

type CartState = {
  lines: CartLine[];
  setLines: (lines: CartLine[]) => void;
};

export const useCartStore = create<CartState>((set) => ({
  lines: [],
  setLines: (lines) => set({ lines }),
}));
