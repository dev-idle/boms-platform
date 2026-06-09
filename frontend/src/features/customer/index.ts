/**
 * Customer feature — cart, checkout, and orders (session required).
 *
 * Public catalog browse lives in `features/catalog`.
 */
export {
  addCartItem,
  applyCartDiscount,
  checkoutCart,
  getCart,
  getOrder,
  listOrders,
  removeCartDiscount,
  removeCartItem,
  updateCartItem,
} from "./api";
export {
  AddToCartButton,
  CartView,
  OrderDetail,
  OrderList,
  ProductPurchaseActions,
} from "./components";
export {
  customerQueryKeys,
  useAddCartItem,
  useApplyCartDiscount,
  useCart,
  useCheckoutCart,
  useOrder,
  useOrders,
  useRemoveCartDiscount,
  useRemoveCartItem,
  useUpdateCartItem,
} from "./hooks";
export {
  cartSchema,
  orderSchema,
  ordersListFilterSchema,
} from "./schemas";
export type {
  AddCartItemInput,
  ApplyCartDiscountInput,
  Cart,
  Order,
  OrdersListFilterInput,
  OrdersListResult,
  UpdateCartItemInput,
} from "./schemas";
