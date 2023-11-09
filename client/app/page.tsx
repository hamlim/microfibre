import { Button } from "@recipes/button";
import { Container } from "@recipes/container";
import { Heading } from "@recipes/heading";
import { Input } from "@recipes/input";
import { Label } from "@recipes/label";
import { BaseLink } from "@recipes/link";
import { Stack } from "@recipes/stack";
import { Textarea } from "@recipes/textarea";

export default function Home() {
  async function submit(formData: FormData) {
    "use server";

    let body = formData.get("body");
    let location = formData.get("location");

    if (!body || typeof body !== "string") {
      throw new Error("Missing body in status update!");
    }

    let payload: { body: string; location?: string } = { body: body as string };
    if (typeof location === "string" && location.length) {
      payload.location = location;
    }

    let endpoint;
    if (process.env.NODE_ENV === "development") {
      endpoint = "http://127.0.0.1:8080/v1/create";
    } else {
      endpoint = "https://microfibre-v1.fly.dev/v1/create";
    }

    await fetch(endpoint, {
      headers: new Headers({
        "secret-token": "yoloswag",
        "api-version": "v1",
      }),
      method: "POST",
      body: JSON.stringify(payload),
    });
    // redirect("/");
  }
  return (
    <main className="p-10">
      <Container>
        <Stack gap={10}>
          <header>
            <Heading is="h1">microfibre</Heading>
          </header>
          <article className="max-w-[70ch] min-w-full md:min-w-[70ch] mx-auto">
            <Stack gap={10}>
              <Heading is="h3">Add Update:</Heading>
              <form action={submit}>
                <Stack gap={10}>
                  <div className="grid w-full items-center gap-4">
                    <Label htmlFor="body">Update</Label>
                    <Textarea id="body" name="body" placeholder="What have you been up to?" required />
                  </div>
                  <div className="grid w-full items-center gap-4">
                    <Label htmlFor="location">Location</Label>
                    <Input type="text" id="location" name="location" placeholder="Location" />
                  </div>
                  <Button
                    variant="outline"
                    type="submit"
                  >
                    Submit
                  </Button>
                </Stack>
              </form>
            </Stack>
          </article>
          <footer>
            Made by <BaseLink href="https://matthamlin.me">Matt Hamlin</BaseLink>
          </footer>
        </Stack>
      </Container>
    </main>
  );
}
