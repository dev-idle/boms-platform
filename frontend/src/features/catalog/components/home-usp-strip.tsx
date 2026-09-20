/** Ordinals carry the rhythm here — the strip deliberately has no icons. */
const USP_ITEMS = [
  {
    title: "Baked fresh daily",
    text: "Small batches from our ovens every morning.",
  },
  {
    title: "Order ahead",
    text: "Reserve your pickup time at checkout.",
  },
  {
    title: "Handcrafted",
    text: "Patisserie made by hand, no shortcuts.",
  },
  {
    title: "Easy pickup",
    text: "Skip the line — grab and go in store.",
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
