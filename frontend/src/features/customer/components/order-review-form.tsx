"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { FieldControl } from "@/components/ui/field-control";
import { Form, FormControl, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { isApiError } from "@/lib/errors";
import { REVIEW_COMMENT_MAX_LENGTH } from "@/lib/schemas/review";
import { applyFormFieldErrors } from "@/lib/validation";

import { useCreateReview } from "../hooks";
import { reviewInputSchema, type ReviewInput } from "../schemas";
import { RatingInput } from "./rating-input";

type OrderReviewFormProps = {
  orderId: string;
  productId: string;
  productName: string;
};

const FIELDS = ["rating", "comment"] as const;

/** The customer's rating of one product they picked up and, if they like, what they thought of it. */
export function OrderReviewForm({ orderId, productId, productName }: OrderReviewFormProps) {
  const form = useForm<ReviewInput>({
    resolver: zodResolver(reviewInputSchema),
    defaultValues: { product_id: productId, rating: 0, comment: "" },
  });
  const create = useCreateReview(orderId);

  function onSubmit(values: ReviewInput): void {
    create.mutate(values, {
      onError: (error) => {
        if (isApiError(error) && error.hasValidationDetails()) {
          applyFormFieldErrors(form, error.details!, FIELDS);
        }
      },
    });
  }

  return (
    <Form {...form}>
      <form
        aria-label={`Review of ${productName}`}
        className="order-review-form"
        method="post"
        noValidate
        onSubmit={form.handleSubmit(onSubmit)}
      >
        <FormField
          control={form.control}
          name="rating"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <fieldset className="order-review-form__rating">
                  <legend className="text-form-label">Rating</legend>
                  <RatingInput
                    disabled={create.isPending}
                    name={`rating-${productId}`}
                    value={field.value}
                    onChange={field.onChange}
                  />
                </fieldset>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="comment"
          render={({ field }) => (
            <FormItem>
              <FieldControl hint={`Up to ${REVIEW_COMMENT_MAX_LENGTH} characters.`} label="Comment" optional>
                <textarea
                  className="field-chrome text-form-input"
                  disabled={create.isPending}
                  placeholder={`What did you think of the ${productName}?`}
                  rows={3}
                  {...field}
                />
              </FieldControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button aria-busy={create.isPending || undefined} disabled={create.isPending} type="submit" variant="outline">
          {create.isPending ? "Posting…" : "Post review"}
        </Button>
      </form>
    </Form>
  );
}
