import type { FeedImage, ImagePage } from "@/types";

type ListImagesOptions = {
  limit?: number;
  cursor?: string;
  tags?: string[];
};

export async function listImages(
  options: ListImagesOptions = {},
  signal?: AbortSignal,
): Promise<ImagePage> {
  const params = new URLSearchParams();
  if (options.limit !== undefined) params.set("limit", String(options.limit));
  if (options.cursor) params.set("cursor", options.cursor);
  for (const tag of options.tags ?? []) params.append("tag", tag);

  const query = params.size > 0 ? `?${params.toString()}` : "";
  const response = await fetch(`/api/images${query}`, {
    headers: { Accept: "application/json" },
    signal,
  });
  if (!response.ok) throw await responseError(response);
  return (await response.json()) as ImagePage;
}

export type UploadInput = {
  file: File;
  title: string;
  tags: string[];
};

export async function uploadImage(input: UploadInput, signal?: AbortSignal): Promise<FeedImage> {
  const body = new FormData();
  body.set("title", input.title);
  if (input.tags.length > 0) body.set("tags", input.tags.join(","));
  body.set("file", input.file);

  const response = await fetch("/uploads", { method: "POST", body, signal });
  if (!response.ok) throw await responseError(response);
  return (await response.json()) as FeedImage;
}

export function websocketUrl(): string {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}/api/ws`;
}

export function describeError(cause: unknown): string {
  return cause instanceof Error ? cause.message : "Something went wrong";
}

async function responseError(response: Response): Promise<Error> {
  let message = `The request failed with status ${response.status}.`;

  try {
    const body = (await response.json()) as { error?: { message?: unknown } };
    if (typeof body.error?.message === "string") message = body.error.message;
  } catch {
    // A non-JSON body (a proxy error page, say) leaves the default in place.
  }

  return new Error(message);
}
