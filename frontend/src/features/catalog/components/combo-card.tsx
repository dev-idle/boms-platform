import type { ReactNode } from "react";

import type { CatalogCombo } from "@/lib/schemas/catalog";
import { plural } from "@/lib/utils";
import { formatPriceCents } from "@/lib/validation/catalog";

import { comboSavingsCents, formatComboEndsAt } from "../lib/combo-pricing";
import { CatalogTile } from "./catalog-tile";

type ComboCardProps = {
  combo: CatalogCombo;
  renderPurchaseActions?: (comboId: string) => ReactNode;
};

/** A bundle in the product card's anatomy: photo, identity, then the transaction. */
export function ComboCard({ combo, renderPurchaseActions }: ComboCardProps) {
  const savingsCents = comboSavingsCents(combo);
  const itemCount = combo.items.reduce((sum, item) => sum + item.quantity, 0);

  return (
    <article className="catalog-combo-card">
      <CatalogTile name={combo.name} scale="card" url={combo.image_url} />
      <div className="catalog-product-card__body">
        <p className="catalog-product-card__eyebrow">Combo · {plural(itemCount, "item")}</p>
        <h3 className="catalog-product-card__name">{combo.name}</h3>
        <p className="catalog-combo-card__items">
          {combo.items.map((item) => `${item.product_name} ×${item.quantity}`).join(" · ")}
        </p>
        {combo.sold_out_today ? <p className="catalog-sold-out">Sold out today</p> : null}
        <div className="catalog-product-card__footer">
          <p className="catalog-combo-card__price">
            <span className="text-price">{formatPriceCents(combo.price_cents)}</span>
            {savingsCents > 0 ? (
              <span className="catalog-combo-card__savings">Save {formatPriceCents(savingsCents)}</span>
            ) : null}
          </p>
          <p className="catalog-combo-card__ends">Until {formatComboEndsAt(combo.ends_at)}</p>
        </div>
        {renderPurchaseActions ? (
          <div className="catalog-combo-card__actions">{renderPurchaseActions(combo.id)}</div>
        ) : null}
      </div>
    </article>
  );
}
