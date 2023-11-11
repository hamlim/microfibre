import { Container } from "@recipes/container";
import { Heading } from "@recipes/heading";
import { BaseLink } from "@recipes/link";
import { Stack } from "@recipes/stack";
import { submit } from "./action";
import { Form } from "./Form";

export default function Home() {
  return (
    <main className="p-10">
      <Container>
        <Stack gap={10}>
          <header className="text-center">
            <Heading is="h1">microfibre</Heading>
          </header>
          <article className="max-w-[70ch] min-w-full md:min-w-[70ch] mx-auto">
            <Stack gap={10}>
              <Heading is="h3">New Post:</Heading>
              <Form action={submit} />
            </Stack>
          </article>
          <footer className="text-center">
            Made by <BaseLink href="https://matthamlin.me">Matt Hamlin</BaseLink>
          </footer>
        </Stack>
      </Container>
    </main>
  );
}
