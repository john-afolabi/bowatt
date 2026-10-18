export const MAX_UPLOAD_BYTES = 10 * 1024 * 1024;
export const MAX_TITLE_LENGTH = 120;
export const MAX_TAGS = 10;
export const MAX_TAG_LENGTH = 24;
export const ACCEPTED_MIME_TYPES = ["image/jpeg", "image/png", "image/gif"] as const;
export const ACCEPT_ATTRIBUTE = ACCEPTED_MIME_TYPES.join(",");

export type UploadValues = {
  file: File | null;
  title: string;
  tagsInput: string;
};

export type UploadErrors = {
  file?: string;
  title?: string;
  tags?: string;
};

export function parseTags(input: string): string[] {
  const tags: string[] = [];
  for (const raw of input.split(",")) {
    const tag = raw
      .trim()
      .toLowerCase()
      .replace(/[\s_./]+/g, "-")
      .replace(/[^a-z0-9-]/g, "")
      .replace(/-{2,}/g, "-")
      .replace(/^-+|-+$/g, "");

    if (tag && !tags.includes(tag)) tags.push(tag);
  }
  return tags.sort();
}

export function validateUpload(values: UploadValues): UploadErrors {
  const errors: UploadErrors = {};
  const title = values.title.trim();
  const tags = parseTags(values.tagsInput);

  if (!values.file) {
    errors.file = "Choose an image to upload.";
  } else if (!(ACCEPTED_MIME_TYPES as readonly string[]).includes(values.file.type)) {
    errors.file = "Only JPEG, PNG and GIF images are supported.";
  } else if (values.file.size === 0) {
    errors.file = "That file is empty.";
  } else if (values.file.size > MAX_UPLOAD_BYTES) {
    errors.file = `Images must be at most ${MAX_UPLOAD_BYTES / (1024 * 1024)} MB.`;
  }

  if (!title) {
    errors.title = "Give the image a title.";
  } else if (title.length > MAX_TITLE_LENGTH) {
    errors.title = `Titles are at most ${MAX_TITLE_LENGTH} characters.`;
  }

  if (values.tagsInput.trim() && tags.length === 0) {
    errors.tags = "Those tags have no letters or digits in them.";
  } else if (tags.length > MAX_TAGS) {
    errors.tags = `Use at most ${MAX_TAGS} tags.`;
  } else if (tags.some((tag) => tag.length > MAX_TAG_LENGTH)) {
    errors.tags = `Tags are at most ${MAX_TAG_LENGTH} characters.`;
  }

  return errors;
}

export function hasErrors(errors: UploadErrors): boolean {
  return Object.values(errors).some((message) => message !== undefined);
}
