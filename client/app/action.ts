"use server";

import { PutObjectCommand, S3Client } from "@aws-sdk/client-s3";
import { getSignedUrl } from "@aws-sdk/s3-request-presigner";

type Media = {
  url: string;
  type: "image" | "video";
  title: string;
};

type Payload = {
  body: string;
  created_timezone: string;
  location?: string;
  media?: Array<Media>;
};

const S3 = new S3Client({
  region: "auto",
  endpoint: `https://${process.env.R2_ACCOUNT_ID}.r2.cloudflarestorage.com`,
  credentials: {
    accessKeyId: process.env.R2_UPLOAD_KEY as string,
    secretAccessKey: process.env.R2_UPLOAD_SECRET as string,
  },
});

export async function submit(formData: FormData) {
  "use server";

  let body = formData.get("body");
  let location = formData.get("location");
  let timezone = formData.get("timezone");
  let media = formData.getAll("media") as Array<File>;

  let hasFiles = false;
  if (media && media.length && !(media.length === 1 && media[0].name === "undefined")) {
    hasFiles = true;
  }

  if (hasFiles) {
    let fileUploads = media.map(file => {
      return getSignedUrl(
        S3,
        new PutObjectCommand({ Bucket: process.env.R2_UPLOAD_BUCKET, Key: file.name }),
        {
          expiresIn: 3600,
        },
      ).then(url =>
        fetch(url, {
          method: "PUT",
          body: file,
        })
      );
    });

    try {
      let res = await Promise.allSettled(fileUploads);
      console.log(res);
    } catch (e) {
      console.log(
        `Failed to upload files: number of files: ${media.length}, filenames: ${
          media.map(f => f.name).join(", ")
        }, error: ${(e as unknown as Error).toString()}`,
      );
    }
  }

  if (!body || typeof body !== "string") {
    throw new Error("Missing body in status update!");
  }

  let payload: Payload = {
    body: body as string,
    created_timezone: timezone as string,
  };
  if (typeof location === "string" && location.length) {
    payload.location = location;
  }
  // @TODO!
  // if (hasFiles) {
  //   payload.media = media.map()
  // }

  let endpoint;
  if (process.env.NODE_ENV === "development") {
    endpoint = "http://127.0.0.1:8787/v1/post";
  } else {
    endpoint = "https://microfibre-api.mhamlin.workers.dev/v1/post";
  }

  await fetch(endpoint, {
    headers: new Headers({
      "x-auth-token": "yolo-swag",
    }),
    method: "POST",
    body: JSON.stringify(payload),
  });
}
