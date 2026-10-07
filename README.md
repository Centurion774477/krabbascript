# KrabbaScript

![GitHub License](https://img.shields.io/github/license/khytryy/krabbascript)
![GitHub top language](https://img.shields.io/github/languages/top/khytryy/krabbascript?logo=go&label=)

KrabbaScript is a simple yet powerful programming language, taking inspiration from C, Lua and Python. It is compiled, statically typed and type-safe, requiring a type for every variable declaration.

> [!CAUTION]
> This project is still W.I.P and some stuff are not finished. Check out our [Discord](https://discord.gg/MQT4YgEYvn) for news and updates

## Getting started

## Building the project
You can simply build the compiler by running `go build`

```bash
go build
```

## Syntax

This is a brief overview of the docs. Don't expect an in-depth explanation with jargon and whatnot, but you'll get a sense of Krabbascript's style. Let's begin!

As with most statically typed and compiled languages, Krabbascript enforces semicolons to separate expressions.

### Variables

Krabbascript has two options for variables. Naturally, you can have variables, or, you can have values. What's the difference? Variables are mutable -- they can be mutated and reassigned -- values are not. You can declare a variable or value using `var` or `val` respectively, the name of the variable, and then a type. Here's an example:

```
var butterfly: I32 = 1704;
```

Great. If you wanted to create a value, it would be the same approach but with `val`:

```
val butterfly: I32 = 96;
```

If you try to mutate or reassign a value, you'll be hit with an error.

Quick note on variables: mutate them all you want, but you can't touch the type. `butterfly` will always be an I32 for the rest of eternity. It's fate is sealed.

### Control structures

Once again, like most statically typed and compiled languages, Krabbascript uses curly brackets for its blocks. 
However, we won't make you use parens for your conditions. Your welcome.

Krabbascript has a solid arrangement of control structures for you. We've got if-elsif-else blocks, while loops, repeat loops, for loops and when blocks.

### if-elsif-else blocks

```
if breakfast {
  drinkFrenchPressCoffeeScript();
} elsif lunch {
  eatSwedishMeatballs();
} else {
  eatCrab();
}
```

### while loops

Once again, no parens needed.

```
while railsIsUnpopular() {
  rantAboutWhyRailsIsGreat();
}
```

### repeat loops

In these, you loop now and declare a condition later.

```
repeat {
  rantAboutWhyRailsIsGreat();
} until railsIsTrending
```

### when blocks

This is one of the more unique parts of Krabbascript. This is like a switch in nature, but with a slightly different approach.

```
when language {
  == "Elixir" {
    celebrate();
  }
  == "Haskell" {
    complain();
  }
}
```

### for loops

The for loops here are closer to foreach loops, and the design was heavily inspired by Lua.

```
for index, language in languages {
    when language {
    == "Elixir" {
      celebrate();
    }
    == "Haskell" {
      complain();
    }
  }
  std.print(index);
}
```

If you don't want that pesky index variable, feel free to use an underscore in its place. Like:

```
for _, variable in arrays
```
