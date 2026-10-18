import { useCallback, useEffect, useRef, useState } from "react";
import type { DragEvent, FormEvent } from "react";
import { ImagePlus, Loader2, Upload, X } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { describeError, uploadImage } from "@/lib/api";
import { cn } from "@/lib/utils";
import {
  ACCEPT_ATTRIBUTE,
  hasErrors,
  MAX_TAGS,
  parseTags,
  validateUpload,
  type UploadErrors,
} from "@/lib/validation";
import type { FeedImage } from "@/types";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onUploaded: (image: FeedImage) => void;
};

export function UploadDialog({ open, onOpenChange, onUploaded }: Props) {
  const [file, setFile] = useState<File | null>(null);
  const [title, setTitle] = useState("");
  const [tagsInput, setTagsInput] = useState("");
  const [errors, setErrors] = useState<UploadErrors>({});
  const [submitError, setSubmitError] = useState<string | undefined>(undefined);
  const [submitting, setSubmitting] = useState(false);
  const [dragging, setDragging] = useState(false);
  const [previewUrl, setPreviewUrl] = useState<string | undefined>(undefined);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    if (!open) return;
    setFile(null);
    setTitle("");
    setTagsInput("");
    setErrors({});
    setSubmitError(undefined);
    setSubmitting(false);
  }, [open]);

  useEffect(() => {
    if (!file) {
      setPreviewUrl(undefined);
      return;
    }
    const url = URL.createObjectURL(file);
    setPreviewUrl(url);
    return () => URL.revokeObjectURL(url);
  }, [file]);

  const chooseFile = useCallback((candidate: File | undefined) => {
    if (!candidate) return;
    setFile(candidate);
    setErrors((current) => ({ ...current, file: undefined }));
  }, []);

  const handleDrop = (event: DragEvent<HTMLLabelElement>) => {
    event.preventDefault();
    setDragging(false);
    chooseFile(event.dataTransfer.files[0]);
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const validation = validateUpload({ file, title, tagsInput });
    setErrors(validation);
    if (hasErrors(validation) || !file) return;

    setSubmitting(true);
    setSubmitError(undefined);
    try {
      const image = await uploadImage({ file, title: title.trim(), tags: parseTags(tagsInput) });
      onUploaded(image);
      onOpenChange(false);
    } catch (cause) {
      setSubmitError(describeError(cause));
    } finally {
      setSubmitting(false);
    }
  };

  const previewTags = parseTags(tagsInput);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Upload an image</DialogTitle>
        </DialogHeader>

        <form className="space-y-4" onSubmit={handleSubmit} noValidate>
          <div className="space-y-2">
            <Label htmlFor="upload-file">Image</Label>
            <label
              htmlFor="upload-file"
              onDragOver={(event) => {
                event.preventDefault();
                setDragging(true);
              }}
              onDragLeave={() => setDragging(false)}
              onDrop={handleDrop}
              className={cn(
                "flex cursor-pointer flex-col items-center justify-center gap-2 rounded-lg border border-dashed p-4 text-center text-sm transition-colors",
                dragging ? "border-primary bg-accent/60" : "border-input hover:bg-accent/40",
                errors.file && "border-destructive",
              )}
            >
              {previewUrl ? (
                <img
                  src={previewUrl}
                  alt="Selected image preview"
                  className="max-h-40 w-auto rounded-md object-contain"
                />
              ) : (
                <ImagePlus aria-hidden className="size-6 text-muted-foreground" />
              )}
              <span className="text-muted-foreground">
                {file ? "Choose a different image" : "Drop an image here, or click to browse"}
              </span>
              <span className="text-xs text-muted-foreground">JPEG, PNG or GIF, up to 10 MB</span>
            </label>
            <input
              ref={fileInputRef}
              id="upload-file"
              name="file"
              type="file"
              accept={ACCEPT_ATTRIBUTE}
              className="sr-only"
              aria-invalid={errors.file ? true : undefined}
              aria-describedby={errors.file ? "upload-file-error" : undefined}
              onChange={(event) => chooseFile(event.target.files?.[0])}
            />

            {file && (
              <div className="flex items-center justify-between gap-2 text-xs text-muted-foreground">
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setFile(null);
                    if (fileInputRef.current) fileInputRef.current.value = "";
                  }}
                >
                  <X aria-hidden />
                  Remove
                </Button>
              </div>
            )}

            {errors.file && (
              <p id="upload-file-error" className="text-xs text-destructive">
                {errors.file}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="upload-title">Title</Label>
            <Input
              id="upload-title"
              name="title"
              value={title}
              maxLength={160}
              aria-invalid={errors.title ? true : undefined}
              aria-describedby={errors.title ? "upload-title-error" : undefined}
              onChange={(event) => {
                setTitle(event.target.value);
                setErrors((current) => ({ ...current, title: undefined }));
              }}
            />
            {errors.title && (
              <p id="upload-title-error" className="text-xs text-destructive">
                {errors.title}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="upload-tags">Tags</Label>
            <Input
              id="upload-tags"
              name="tags"
              value={tagsInput}
              aria-invalid={errors.tags ? true : undefined}
              aria-describedby={errors.tags ? "upload-tags-error" : "upload-tags-hint"}
              onChange={(event) => {
                setTagsInput(event.target.value);
                setErrors((current) => ({ ...current, tags: undefined }));
              }}
            />
            <p id="upload-tags-hint" className="text-xs text-muted-foreground">
              Comma separated, up to {MAX_TAGS}. Letters, digits and dashes.
            </p>

            {previewTags.length > 0 && (
              <div className="flex flex-wrap gap-1">
                {previewTags.map((tag) => (
                  <Badge key={tag} variant="secondary">
                    {tag}
                  </Badge>
                ))}
              </div>
            )}

            {errors.tags && (
              <p id="upload-tags-error" className="text-xs text-destructive">
                {errors.tags}
              </p>
            )}
          </div>

          {submitError && (
            <p
              role="alert"
              className="rounded-md border border-destructive/40 bg-destructive/10 p-2 text-xs text-destructive"
            >
              {submitError}
            </p>
          )}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={submitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? (
                <Loader2 aria-hidden className="animate-spin" />
              ) : (
                <Upload aria-hidden />
              )}
              {submitting ? "Uploading…" : "Upload"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
