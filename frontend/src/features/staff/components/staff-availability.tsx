"use client";

import { useState } from "react";

import { DashboardFilterGroup } from "@/components/ui/dashboard-filter-group";
import { DashboardTableActionButton } from "@/components/ui/dashboard-table-action-button";
import { DashboardTableRowActions } from "@/components/ui/dashboard-table-actions";
import { DashboardTablePagination } from "@/components/ui/dashboard-table-pagination";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { StatusPill } from "@/components/ui/status-pill";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { STATION_LABEL } from "@/lib/schemas/ticket";

import { useSetProductSoldOut, useStaffProducts } from "../hooks";

const PAGE_SIZE = 20;
const COLUMN_COUNT = 5;

const FILTERS: Array<{ value: "sold_out" | undefined; label: string }> = [
  { value: undefined, label: "All products" },
  { value: "sold_out", label: "Sold out today" },
];

/** What the counter sells today: a product that runs out is marked so no order collected today takes it. */
export function StaffAvailability() {
  const [page, setPage] = useState(1);
  const [filter, setFilter] = useState<"sold_out" | undefined>(undefined);
  const productsQuery = useStaffProducts({ page, page_size: PAGE_SIZE, sold_out_today: filter === "sold_out" });
  const setSoldOut = useSetProductSoldOut();
  const { initialLoading, refetching } = getQuerySurface(productsQuery);
  const products = productsQuery.data?.products ?? [];
  const pagination = productsQuery.data?.pagination;

  return (
    <div className="dashboard-page-body">
      <div className="db-table-filters db-table-filters--end">
        <DashboardFilterGroup
          aria-label="Filter by availability"
          onChange={(next) => {
            setFilter(next);
            setPage(1);
          }}
          options={FILTERS}
          value={filter}
        />
      </div>

      <DashboardTableWrap refetching={refetching}>
        <table className="db-table db-table--availability">
          <colgroup>
            <col />
            <col />
            <col className="db-table-col-station" />
            <col className="db-table-col-status" />
            <col className="db-table-col-actions" />
          </colgroup>
          <thead>
            <tr>
              <th>Product</th>
              <th>Category</th>
              <th>Station</th>
              <th className="db-table-status">Status</th>
              <th className="db-table-detail">Actions</th>
            </tr>
          </thead>
          <tbody>
            <DashboardTableStateRows
              columnCount={COLUMN_COUNT}
              emptyFilteredMessage="Nothing is sold out today."
              entityLabel="products"
              hasActiveFilter={filter !== undefined}
              isEmpty={products.length === 0}
              isError={productsQuery.isError}
              initialLoading={initialLoading}
            />
            {!initialLoading && !productsQuery.isError
              ? products.map((product) => {
                  const updating = setSoldOut.isPending && setSoldOut.variables?.productId === product.id;
                  return (
                    <tr key={product.id}>
                      <td className="db-table-cell-primary">{product.name}</td>
                      <td className="text-muted">{product.category_name}</td>
                      <td className="db-table-cell-station">{STATION_LABEL[product.station]}</td>
                      <td className="db-table-status">
                        <StatusPill
                          label={product.sold_out_today ? "Sold out today" : "Available"}
                          variant={product.sold_out_today ? "completed" : "ready"}
                        />
                      </td>
                      <td className="db-table-detail">
                        <DashboardTableRowActions>
                          <DashboardTableActionButton
                            blockedReason={updating ? "Updating the product" : undefined}
                            label={`${product.sold_out_today ? "Bring back" : "Mark sold out"} ${product.name}`}
                            onClick={() => setSoldOut.mutate({ productId: product.id, soldOut: !product.sold_out_today })}
                            text={product.sold_out_today ? "Bring back" : "Mark sold out"}
                            tone={product.sold_out_today ? "accent" : "warning"}
                          />
                        </DashboardTableRowActions>
                      </td>
                    </tr>
                  );
                })
              : null}
          </tbody>
        </table>
        <DashboardTablePagination
          disabled={refetching}
          itemLabel="products"
          onPageChange={setPage}
          page={pagination?.page ?? page}
          pageSize={pagination?.page_size ?? PAGE_SIZE}
          totalItems={pagination?.total ?? products.length}
          totalPages={pagination?.total_pages ?? 1}
        />
      </DashboardTableWrap>
    </div>
  );
}
