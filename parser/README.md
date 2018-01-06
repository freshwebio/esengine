# esengine parser ES8(ES2017)

The parser package for esengine which handles lexing and parsing
and generating abstract syntax trees for given

The reference used for the grammars is the one found here:
https://www.ecma-international.org/publications/files/ECMA-ST/Ecma-262.pdf

When it comes to lexing and unicode characters the bits not covered in the ES8 Spec
(Such as the entire set of Z classified characters for white space)
are taken from the following source:
http://www.fileformat.info/info/unicode

## ECMAScript 8 Grammar

The grammar is ported from the ECMAScript specification to a YAML
file which uses the following conventions.

<[A-Za-z]+> represents a non-terminal symbol in the grammar.
Non-terminal symbols can take the following form:
```
<\w+>
```
```
<\w+>:
  params: [\w+, ...]
  rhs: [[{Terminal|NonTerminal}, ...], ...]
```
The following provide the representations of a non-terminal symbol in the right hand side
of a production:
```
<\w+>
```
```
<\w+>:
  params:
    passthrough: [{Param}, ...]
    optional: true|false
    conditions: [{Param}, ...]
```

Terminal symbols can also have conditions such as `[~Yield] yield` in the ECMAScript specification
which translates to `yield: { conditions: [~Yield] }` in our YAML representation of the grammar where
[] represents a list in this case.

<\*[A-Za-z]+\*> represent custom operations that should occur when a certain position in a production is reached.
The <\*Lookahead\*> operation should take an exclude parameter with a list of terminal or non-terminal symbols
that the next token cannot be.

The <\*Conditional\*> operation represents a condition applied to a right hand side alternative consisting of multiple
terminal and non-terminal symbols. The condition operation takes a `params: { conditions: [...]}` mapping
and a `parts: []`.

<![A-Za-z]+!> represents a placeholder where anything but one or more of the given terminal symbol will follow.
