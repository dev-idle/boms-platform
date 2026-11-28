/** Ordinals carry the rhythm here — the strip deliberately has no icons. */
const USP_ITEMS = [
  {
    title: "Order online",
    text: "Pay with PayPal to complete your order.",
  },
  {
    title: "Reserve your slot",
    text: "Choose your pickup time at checkout.",
  },
  {
    title: "Made to order",
    text: "Custom cakes in your size, flavor, filling and decoration.",
  },
  {
    title: "Pick up with a code",
    text: "Give your 4-digit pickup code at the counter.",
  },
] as const;

export function HomeUspStrip() {
  return (
    <section aria-label="Why order with us" className="storefront-usp-section">
      <div className="storefront-container">
        <ul className="storefront-usp">
          {USP_ITEMS.map((item, index) => (
            <li key={item.title} className="storefront-usp__item">
              <p className="storefront-usp__ordinal">
                {String(index + 1).padStart(2, "0")}
              </p>
              <p className="storefront-usp__title">{item.title}</p>
              <p className="storefront-usp__text">{item.text}</p>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
}
