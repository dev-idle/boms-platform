import type { ReactNode } from "react";

import { formatPriceCents } from "@/lib/validation/catalog";

import {
  comboSavingsCents,
  formatComboEndsAt,
} from "../lib/combo-pricing";

import type { CatalogCombo } from "@/lib/schemas/catalog";

type ComboCardProps = {
  combo: CatalogCombo;
  renderPurchaseActions?: (comboId: string) => ReactNode;
};

export function ComboCard({ combo, renderPurchaseActions }: ComboCardProps) {
  const savingsCents = comboSavingsCents(combo);
  const itemCount = combo.items.reduce((sum, item) => sum + item.quantity, 0);

  return (
    <article className="catalog-combo-card">
      <header className="catalog-combo-card__header">
        <div className="catalog-combo-card__badges">
          <span className="catalog-combo-card__badge">Bundle</span>
          <span className="catalog-combo-card__ends">
            Ends {formatComboEndsAt(combo.ends_at)}
          </span>
        </div>
        <h3 className="catalog-combo-card__title">{combo.name}</h3>
        <p className="catalog-combo-card__meta">
          {itemCount} {itemCount === 1 ? "item" : "items"} included
        </p>
      </header>

      <ul className="catalog-combo-card__items">
        {combo.items.map((item) => (
          <li key={`${combo.id}-${item.product_id}`}>
            <span className="catalog-combo-card__item-line">
              <span className="catalog-combo-card__item-name">
                {item.product_name}
              </span>
              <span className="catalog-combo-card__qty">×{item.quantity}</span>
            </span>
          </li>
        ))}
      </ul>

      <footer className="catalog-combo-card__footer">
        <div className="catalog-combo-card__pricing">
          <p className="catalog-combo-card__price text-price">
            {formatPriceCents(combo.price_cents)}
          </p>
          {savingsCents > 0 ? (
            <p className="catalog-combo-card__savings">
              Save {formatPriceCents(savingsCents)}
            </p>
          ) : null}
        </div>
        {renderPurchaseActions ? (
          <div className="catalog-combo-card__actions">
            {renderPurchaseActions(combo.id)}
          </div>
        ) : null}
      </footer>
    </article>
  );
}
