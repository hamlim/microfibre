import { cn } from "@recipes/cn";

interface Props extends React.HTMLProps<HTMLDivElement> {
}

export function Container(props: Props) {
  return <div {...props} className={cn("container", props.className)} />;
}
