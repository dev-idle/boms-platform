import { PRODUCT_OPTION_GROUP_LABEL } from "@/lib/schemas/catalog";
import type { Customization } from "@/lib/schemas/order";

type CustomizationSummaryProps = {
  customization: Customization;
};

/** How a customer configured a custom item: each option by group, the cake message and the reference photo. */
export function CustomizationSummary({ customization }: CustomizationSummaryProps) {
  const { options, message, reference_image_url: photo } = customization;
  return (
    <ul className="customization-summary">
      {options.map((option) => (
        <li key={option.group}>
          <span className="customization-summary__label">{PRODUCT_OPTION_GROUP_LABEL[option.group]}</span>{" "}
          {option.label}
        </li>
      ))}
      {message ? (
        <li>
          <span className="customization-summary__label">Message</span> “{message}”
        </li>
      ) : null}
      {photo ? (
        <li>
          <a className="customization-summary__photo-link" href={photo} rel="noopener noreferrer" target="_blank">
            {/* eslint-disable-next-line @next/next/no-img-element -- the customer's own Cloudinary upload */}
            <img alt="Reference photo" className="customization-summary__photo" src={photo} />
            <span className="sr-only"> (opens in a new tab)</span>
          </a>
        </li>
      ) : null}
    </ul>
  );
}
