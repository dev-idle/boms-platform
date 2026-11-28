import { catalogProductImageUrl } from "@/lib/cloudinary/config";
import { cn } from "@/lib/utils";

import { catalogItemInitial } from "../lib/catalog-item-initial";

/** Where the tile sits, which sets its size and the weight of its initial. */
type CatalogTileScale = "hero" | "card" | "row" | "search";

/** The width each scale fetches, in CSS pixels. */
const SOURCE_WIDTH: Record<CatalogTileScale, number> = {
  hero: 960,
  card: 720,
  row: 176,
  search: 104,
};

type CatalogTileProps = {
  /** Names the item: its initial is the no-photo state. */
  name: string;
  /** The item's own photo; none shows its initial, never another bake's photo. */
  url: string | null | undefined;
  scale: CatalogTileScale;
  /** The photo's text where it is the page's subject; decorative otherwise. */
  alt?: string;
};

/**
 * A product or combo's photo, or a paper tile carrying its initial as a
 * watermark (02-COMPONENTS §7c) — one treatment at every scale.
 */
export function CatalogTile({ name, url, scale, alt = "" }: CatalogTileProps) {
  const source = catalogProductImageUrl(url, SOURCE_WIDTH[scale]);

  return (
    <span className={cn("catalog-tile", `catalog-tile--${scale}`)}>
      {source ? (
        // eslint-disable-next-line @next/next/no-img-element -- manager-provided catalog URL
        <img alt={alt} className="catalog-tile__image" src={source} />
      ) : (
        <span aria-hidden="true" className="catalog-tile__initial">
          {catalogItemInitial(name)}
        </span>
      )}
    </span>
  );
}
