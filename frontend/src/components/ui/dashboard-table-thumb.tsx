import { catalogItemInitial } from "@/features/catalog";
import { catalogProductImageUrl } from "@/lib/cloudinary/config";

/** Tile width in CSS pixels; the source is fetched at twice that for retina. */
const THUMB_SOURCE_WIDTH = 96;

type DashboardTableThumbProps = {
  /** Names the tile: its initial is the no-photo state (02-COMPONENTS §7c). */
  name: string;
  url: string | null | undefined;
};

/** The 52px lead-column tile of a manager table: the photo, or its monogram. */
export function DashboardTableThumb({ name, url }: DashboardTableThumbProps) {
  const source = catalogProductImageUrl(url, THUMB_SOURCE_WIDTH);

  return (
    <td className="db-table-cell-thumb">
      <span className="db-table-thumb">
        {source ? (
          // eslint-disable-next-line @next/next/no-img-element -- manager-provided catalog URL
          <img
            alt=""
            aria-hidden
            className="db-table-thumb__image"
            src={source}
          />
        ) : (
          <span className="db-table-thumb__initial">
            {catalogItemInitial(name)}
          </span>
        )}
      </span>
    </td>
  );
}
