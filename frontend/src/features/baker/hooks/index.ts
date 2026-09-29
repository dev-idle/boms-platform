"use client";

import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { z } from "zod";
import { toast } from "sonner";

import { isApiError } from "@/lib/errors";
import {
  stationTicketsListFilterSchema,
  ticketChangeMessage,
  type PatchTicketStatusInput,
  type StationTicketsListFilterInput,
} from "@/lib/schemas/ticket";

import {
  getKitchenTicket,
  listKitchenTickets,
  patchKitchenTicketStatus,
} from "../api";
import { bakerQueryKeys } from "./query-options";

const defaultTicketsFilter: StationTicketsListFilterInput = {
  page: 1,
  page_size: 20,
};

function bakerMutationErrorMessage(error: unknown, fallback: string): string {
  if (isApiError(error)) {
    return error.message;
  }
  return fallback;
}

export function useKitchenTickets(
  input: StationTicketsListFilterInput = defaultTicketsFilter,
) {
  const filter = stationTicketsListFilterSchema.parse(input);
  return useQuery({
    queryKey: bakerQueryKeys.tickets(filter),
    queryFn: () => listKitchenTickets(filter),
    placeholderData: keepPreviousData,
  });
}

export function useKitchenTicket(id: string) {
  const isValidId = z.uuid().safeParse(id).success;
  return useQuery({
    queryKey: bakerQueryKeys.ticket(id),
    queryFn: () => getKitchenTicket(id),
    enabled: isValidId,
    retry: false,
  });
}

export function usePatchKitchenTicketStatus(ticketId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: PatchTicketStatusInput) =>
      patchKitchenTicketStatus(ticketId, input),
    onSuccess: (change) => {
      queryClient.invalidateQueries({ queryKey: bakerQueryKeys.ticketsRoot });
      toast.success(ticketChangeMessage(change));
    },
    onError: (error) => {
      toast.error(bakerMutationErrorMessage(error, "Failed to update the ticket"));
    },
  });
}
