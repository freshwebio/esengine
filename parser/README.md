# esengine parser ES8(ES2017)

The parser package for esengine which handles lexing and parsing
and generating abstract syntax trees for given

The reference used for the grammars is the one found here:
https://www.ecma-international.org/publications/files/ECMA-ST/Ecma-262.pdf

When it comes to lexing and unicode characters the bits not covered in the ES8 Spec
(Such as the entire set of Z classified characters for white space)
are taken from the following source:
http://www.fileformat.info/info/unicode
