type UspIconProps = {
  path: string;
};

function UspIcon({ path }: UspIconProps) {
  return (
    <svg
      aria-hidden="true"
      className="storefront-usp__icon"
      fill="none"
      height="22"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.65"
      viewBox="0 0 24 24"
      width="22"
    >
      <path d={path} />
    </svg>
  );
}

const USP_ITEMS = [
  {
    title: "Baked fresh daily",
    text: "Small batches from our ovens every morning.",
    iconPath: "M12 3v3m6.36-.36-2.12 2.12M21 12h-3M5.64 5.64l2.12 2.12M3 12h3m-1 8h14a7 7 0 0 0-14 0Z",
  },
  {
    title: "Order ahead",
    text: "Reserve your pickup time at checkout.",
    iconPath: "M12 8v4l2.5 2.5M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z",
  },
  {
    title: "Handcrafted",
    text: "Patisserie made by hand, no shortcuts.",
    iconPath: "M7 21h10M9 21V11.5L5.5 8A5.5 5.5 0 0 1 12 3a5.5 5.5 0 0 1 6.5 5L15 11.5V21",
  },
  {
    title: "Easy pickup",
    text: "Skip the line — grab and go in store.",
    iconPath: "M6 8h12l1 13H5L6 8Zm3 0a3 3 0 0 1 6 0",
  },
] as const;

export function HomeUspStrip() {
  return (
    <section aria-label="Why order with us" className="border-t border-border bg-bg py-8 sm:py-10">
      <div className="storefront-container">
        <ul className="storefront-usp">
          {USP_ITEMS.map((item) => (
            <li key={item.title} className="storefront-usp__item">
              <UspIcon path={item.iconPath} />
              <div>
                <p className="storefront-usp__title">{item.title}</p>
                <p className="storefront-usp__text">{item.text}</p>
              </div>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
}
