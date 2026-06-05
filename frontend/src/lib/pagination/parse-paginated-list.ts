import { z } from "zod";

const listMetaSchema = z.object({
  request_id: z.string().optional(),
  pagination: z
    .object({
      page: z.number().int().min(1),
      page_size: z.number().int().min(1),
      total: z.number().int().min(0),
      total_pages: z.number().int().min(1).optional(),
    })
    .optional(),
});

export type PaginatedListResult<T> = {
  items: T[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  request_id?: string;
};

export function parsePaginatedList<T>(
  data: T[],
  meta: Record<string, unknown> | undefined,
  fallback: { page: number; page_size: number },
): PaginatedListResult<T> {
  const parsedMeta = listMetaSchema.safeParse(meta);
  const rawPagination = parsedMeta.success ? parsedMeta.data.pagination : undefined;
  const page = rawPagination?.page ?? fallback.page;
  const pageSize = rawPagination?.page_size ?? fallback.page_size;
  const total = rawPagination?.total ?? data.length;
  const totalPages =
    rawPagination?.total_pages ?? Math.max(1, Math.ceil(total / pageSize));

  return {
    items: data,
    pagination: { page, page_size: pageSize, total, total_pages: totalPages },
    request_id: parsedMeta.success ? parsedMeta.data.request_id : undefined,
  };
}
