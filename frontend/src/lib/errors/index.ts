/** Client-side API errors (browser fetch + envelope). */
export {
  ApiError,
  ApiErrorCode,
  apiErrorFromPayload,
  isApiError,
  throwApiErrorFromEnvelope,
} from "./api-error";

/** Server-side DAL errors (RSC / Fiber api-client). */
export { BomsApiError, BomsValidationError } from "./server-api-error";
