[![Build Status](https://travis-ci.org/eaburns/peggy.svg?branch=master)](https://travis-ci.org/eaburns/peggy)

# Introduction

Peggy is a Parsing Expression Grammar
([PEG](https://en.wikipedia.org/wiki/Parsing_expression_grammar))
parser generator.

The generated parser is a
[packrat parser](https://en.wikipedia.org/wiki/Parsing_expression_grammar#Implementing_parsers_from_parsing_expression_grammars).
However, the implementation is somewhat novel (I believe).

# Background

Packrat parsers work by doing a recursive descent on the grammar rules,
backtracking when a rule fails to accept.
To prevent exponential backtracking, a memo table remembers
the parse result for each rule, for each point in the input.
This way when the backtracking encounters a subtree of grammar already tried
it can compute the result in constant time by looking up the memo table
instead of computing the parse again.

Because the memo table, packrat parsers for PEG grammars
parse in time linear in the size of the input
and use memory linear in the size of the input too.
(Note that other common parser generators,
such as yacc for LALR(1) grammars
are linear time in the size of the input
and linear space in the _depth of the parse_,
which can be smaller than the input size.)

A common way to implement the memo table is to use a hash table.
The key is a pair of the grammar rule plus the input position,
and the value is the result (result of any parser actions)
of parsing the keyed rule at the keyed input position
and the number of runes consumed, or whether the parse failed.

A problem that I've found with this approach is that
for grammars that tend to do a lot of backtracking,
a significant amount of time is spent modifying and accessing the memo table.
Hash tables lookups and inserts are expected constant time,
but in the face of much backtracking, the constant time can add up.

In addition, hash tables tend to be implemented with linked structures
which take up additional memory to hold pointers and allocation overhead.
Finally, as they grow large, linked sturctures take more time to scan
by a garbage collector.

I originally implemented Peggy to parse the constructed language
[Lojban](https://mw.lojban.org/papri/Lojban)
(see [johaus](http://github.com/eaburns/johaus)).
My initial hash table based implementation performed very poorly on large texts
because of the issues described above:
profiling showed a singificant amount of time spent
on map accesses and garbage collection scanning,
and memory use was too high to parse some texts (4kb)
on my laptop (8GB ram).

I noticed similar issues with the JavaScript- and Java-based PEG parsers for Lojban.

Peggy takes a different approach that was tuned for this use-case.

## Peggy's approach

Peggy computes the result of a parse in two passes instead of one.
The first pass determines whether the grammar accepts,
and builds a table that tracks for each rule tried at each position:
whether the rule accepted and if so how much input was consumed,
or if it failed, how much input was consumed to the deepest failure.
These values can be stored in an array using only integers.

If the first pass acceptes the input, a second pass can quickly follow the table
to try only rules that accept and compute the result of the actions of the rule.

If the first pass fails to accept, another pass can follow the table
and compute a tree tracking which rules failed at the deepest point of the parse.
These can be used to build precise syntax error messages.

The advantage of Peggy's approach is that
the first pass only performs a single allocation: the table — an array of integers.
Accessing the table is just indexing into an array of intergers,
which is cheaper than most hash table lookups.
Since the array only contains integers and no pointers,
it needn't be scanned by the garbage collector.
And finally, whenever a hash table would be relatively densely populated,
an array can be memory efficient.

For the Lojban grammar, this made the difference
between being able to parse full texts
(a 4KB text that needed >8GB of memory was reduced to needing only 2GB,
and multiple minutes were reduced to mere seconds).

## Disadvantages

There are disadvantages to the Peggy approach:

1) The interface is not as simple to use.
However, I hope that you will not find it too difficult.
See the example in the next section for a fairly short wrapper
that warps the Peggy calls into a single, more typical Go function call.

2) For grammars that do not rely as heavily on the memo table
a hash table could be much more memory efficient.

I would like to expand this list, so please send pull requests
if you have other disadvantages of this approach that should be here.

Now, let's see how to use it.

# Input file format

A Peggy input file is UTF-8 encoded.

A Peggy grammar file consists of a _prelude_ followed by a set of _rules_.
The prelude is valid Go code enclosed between { and }.
This code is emitted at the beginning of the generated parser .go file.
It should begin with a package statement then any imports used by the parser.
Any other valid Go code is also permitted.

After the prelude is a set of _rules_ that define the grammar.
Each rule begins with an _identifier_ that is the name of the rule.
After the name is an optional string giving the rule a human-readable name
and marking it as a _leaf_ rule for error reporting (more below).
After the optional string is the token <-.
Next is the expression that defines the rule.

**Example**
```
A <- "Hello," _ ( "World!" / "世界" )
_ <- ( p:. &{ isUnicodeSpace(p) } )+
```

# Expressions

Expressions define the grammar.
The input to each expression is a sequence of runes.
The expression either accepts or rejects the input.
If the expression accepts, it consumes zero or more runes of input,
and evaluates to a result (a Go value).

The types of expressions, in order of precedence, are:
* Choice
* Action
* Sequence
* Label
* Predicate
* Repetition
* Literal, Code Predicate, Identifier, and Subexpression

## Choice

A choice is a sequence of expressions separated by `/`.
Unlike context free grammars, choices in PEG are ordered.

It is an error if the result types of the subexpressions are not all the same.

**Accepts:**
A choice accepts if any of its expressions accept.

**Consumes:**
A choice consumes the runes consumed by its first accepting subexpression
from left-to-right.

**Result:**
The result of a choice has the type and value of its first accepting subexpression
from left-to-right.

**Example:**
```
A / "Hello" / foo:Bar { return string(foo) }
```

## Sequences

A sequence is two a sequence of expressions separated by whitespace.

**Accepts:**
A sequence accepts if each of its subexpressions accepts
on the input remaining after each preceeding subexpression consumes.

**Consumes:**
The sequence consumes from the input
the sum of the number of runes of all its subexpressions.

***Result:**
It is an error if the type of the result of the first expression
is not the same as the type of the result of the second.

If the first expression is a `string`, the type of the sequence is `string`,
and the result is the concatenation of the results of the expressions.

If the first expression is any non-`string` type, T,
the type of the result of the sequence is `[]T`,
and the result itself is the slice from
`append()`ing the results of the subexpressions.

**Example:**
```
"Hello," Space "World" Punctiation
```

## Labels

A label is an identifier followed by : followed by an expression.

Labels are used to create new identifiers used by actions and code predicates.

**Accepts:**
A label accepts if its subexpression accepts.

**Consumes:**
A label consumes the runs of its subexpression.

**Result:**
The result type and value of a label are that of its subexpression.

**Example:**
```
hello:"Hello" "," Space world:( "World" / "世界" )
```

## Predicates

A predicate is a & or ! operator followed by an expression.

**Accepts:**
A predicate with the operator & accepts if its subexpression accepts.

A predicate with the operator ! accepts if its subexpression dose not accept.

**Consumes:**
Predicatse consume no runes.

**Result:**
The result of a predicate is the empty string.

**Example:**
```
!Keyword [a-ZA-Z_] [a-ZA-Z0-9_]*
```

## Repetition

A repetition is an expression followed by either a *, +, or ? operator.

**Accepts:**
A repetition with an operator * or ? always accepts.

A repetition with the operator + accepts if its subexpression accepts.

**Consumes:**
A repetition with an operator * or + consumes all matches of its subexpression.

A repetition with the operator ? consumes at most one match of its subexpression.

**Result:**
If the type of the subexpression is `string`, the result of a repetition is `string`,
and the value is the consumed runes.

Otherwise, if the type of the subexpression is a type `T`:
* if the operator is * or +, the type of the result is `[]T`
and the value is a slice containing all `append`ed subexpression results.
* if the operatior is ?, the type of the result is `*T`
and the value is a pointer to the subexpression result if it accepted
or `nil`.

**Example:**
```
[a-ZA-Z0-9_]* ":"?
```

## Literals

Literals are String Literals, Character Classes, and Dot.

### String Literals

String literals are lexically the same as
[Go String Literals](https://golang.org/ref/spec#String_literals).

**Accepts:**
A string literal accepts if the next runes of input are exactly those of the string.

**Consumes:**
A stirng literal consumes the matching runes of input.

**Result:**
The result is the `string` of consumed runes.

**Example:**
```
"Hello\nWorld!"
```

### Character Classes

A character class is a sequence of characters
between [ and the next, unescaped occurrence of ].
Escapes are treated as per strings.

Character classes are much like that of common regular expression libraries.

**Accepts:**
A character class accepts if the next rune of input is within the class.

If the first character after the opening [ is a ^,
then the character class's acceptance is negated.

A pair of characters surrounding on either side of a - define a _span_.
the character class will accept any rune with a number (codepoint)
between (and including) the two characters
 It is an error if the first is not smaller than the last.

All other characters in the class are treated as a list of accepted runes.

**Consumes:**
A character class consumes one rune of input.

**Result:**
The result is the `string` of the consumed rune.

**Example:**
```
[a-ZA-Z0-9_]
```

### Dot

The character . is an expression.

**Accepts:**
A dot expression accepts if the input is not empty and the next rune is valid.

**Consumes:**
A dot expression consumes a single rune.

**Result:**
The result is the `string` of the consumed rune.

**Example:**
```
.
```

## Code predicates

A code predicate is an operator & or ! followed by a Go expression enclosed in { and }.
The expression must result in a boolean value,
and must be syntactically valid as the condition of an
[if statement](https://golang.org/ref/spec#If_statements).

Label expressions of the containing rule define identifiers accessible in the Go code.
The value of the identifier is a `string` of the input consumed by the labeled expression.
If the labeled expression has yet to accept at the time the code predicate is evalutade, the string is empty.

**Accepts:**

A code predicate with the operator & accepts if the expression evaluates to `true`.

A code predicate with the operator ! accepts if the expression evaluates to `false`.

**Consumes:**
A code predicate consumes no runes of input.

**Result:**
The result of a code predicate is the empty string.

**Example:**
```
p:. &{ isUnicodeSpace(p) }
```

## Identifiers

Identifiers begin with any unicode letter or _
followed by a sequence of zero or more letters, numbers, or _.
Identifiers name a rule of the grammar.
It is an error if the identifier is not the name of the rule of the grammar.

**Accepts:**
An identifier accepts if its named rule accepts.

**Consumes:**
An identifier consumes the runes of its named rule.

**Result:**
The result of an identifier has the type and value of that of its named rule.

**Example:**
```
HelloWorld <- Hello "," Space World
Hello <- "Hello" / "こんいちは"
World <- "World" / "世界"
Space <- ( p:. &{ isUnicodeSpace(p) } )+
```

## Subexpressions

A subexpression is an expression enclosed between ( and ).
They are primarily used for grouping.

**Accepts:**
A subexpression accepts if its inner expression accepts.

**Consumes:**
A subexpression consumes the runes of its inner expression.

**Result:**
The result type and value of a subexpression are that of its inner expression.

**Example:**
```
"Hello, " ( "World" / "世界" )
```

## Actions

Actions are an expression followed by Go code between { and }.
The Go code must be valid as the
[body of a function](https://golang.org/ref/spec#Block).
The Go code must end in a
[return statement](https://golang.org/ref/spec#Return_statements),
and the returned value must be one of:
* [a type conversion](https://golang.org/ref/spec#Conversions)
* [a type assertion](https://golang.org/ref/spec#Type_assertions)
* [a function literal](https://golang.org/ref/spec#Function_literals)
* [a composite literal](https://golang.org/ref/spec#Composite_literals)
* [an &-composite literal](https://golang.org/ref/spec#Address_operators)
* [an int literal](https://golang.org/ref/spec#Integer_literals)
* [a float literal](https://golang.org/ref/spec#Floating-point_literals)
* [a rune literal](https://golang.org/ref/spec#Rune_literals)
* [a string literal](https://golang.org/ref/spec#String_literals)

Label expressions of the containing rule define identifiers accessible in the Go code.
The value of the identifier is the value of the labeled expression if it accepted.
If the labeled expression has yet to accept at the time the action is evaluated,
the value is the zero value of the corresponding type.

In addition there are several other special identifiers accessable to the code:
* `parser` is a pointer to the Peggy `Parser`.
* `start` is the byte offset in the input at which this expression first accepted.
* `end` is the byte offset in the input just after this expression last accepted.

**Accepts:**
An action accepts if its subexpression accepts.

**Consumes:**
An action consumes the runes of its subexpression.

**Result:**
The result of an action has the type of the last return statement
at the end of the block of Go code.
The value is the value returned by the Go code.

**Example:**
```
hello:("Hello" / "こんいちは") ", " world:("World" / "世界") {
	return HelloWorld{
		Hello: hello,
		World: world,
	}
}
```

# Generated code

The output file path is specified by the `-o` command-line option.

All package-level definitions in the generated begin with a prefix, defaulting to `_`. This default makes the definitions unexported. The prefix can be overridden with the `-p` command-line option.

The generated file has a `Parser` type passed to the various parser functions,
and contains between 2 and 4 of functions for each rule defining
several parser _passes_. The passes are:
1. the _accepts_ pass,
2. the _fail_ pass,
3. optionally the _action_ pass, and
4. optionally the _node_ pass.

A typical flow to use a Peggy-generated parser is to:
* Create a new instance of the `Parser` type on a given input.
* Call the accepts function for the root-level grammar rule.
** If the rule did not accept, there was a syntax error:
	call the fail function of the rule to get an `*peg.Fail` tree,
	and pass that to `peg.SimpleError` to get an `error`
	describing the syntax error.
** If the rule accepted, call the action function of the rule
	to get the result of the parse (an AST, evaluation, whatever),
	or call the node pass to get a `*peg.Node` of the syntax tree.

Here is an example:

```
// Parse returns the AST generated by the grammar rule actions.
func Parse(input string) (AstNode, error) {
	parser := _NewParser(input)
	if pos, perr := _RuleAccepts(parser, 0); pos < 0 {
		_, failTree := _RuleFail(parser, perr)
		return nil, peg.SimpleError(input, failTree)
	}
	// Or, instead call _RuleNode(parser, 0)
	// and return a *peg.Node with the syntax tree.
	_, astRoot := _RuleAction(parser, 0)
	return astRoot, nil
}
```

There are a lot of steps.
This allows advanced uses not described here ☺.
(But see, for example,
[this file](https://github.com/eaburns/johaus/blob/master/parser/error.go)
that showcases how to use the `*peg.Fail` tree to construct more precise error messages).

Now let's see what the generated code for each of the passes looks like in moredetail.

## The Parser type

The `Parser` type is mostly intended to be treated as opaque.
It maintains information about the parse to communicate between the multiple passes.

The `Parser` type will have a field named `data` of type `interface{}`,
which is ignored by the generated code.
This field may be used in code predicates or actions to store auxiliary information.
Such a use is considered advanced, and is not recommended
unless you have a thorough understanding of the generated parser.

## Accepts pass

The accepts pass generates a function for each rule of the grammer with a signature of the form:
```
func <Prefix><RuleName>Accepts(parser *<Prefix>Parser, start int) (deltaPos, deltaErr int)
```

The function determines whether the rule accepts the input
beginning from the byte-offset `start`.
If it accepts `deltaPos` is a non-negative number of bytes accepted.
If it does not accept `deltaErr` is the number of bytes from start
until the last rune of input that could not be consumed.

The primary purpose of the accept pass is to determine
whether the language defined by the grammar accepts the input.
The `Parser` maintains state from the accept pass that enables a subsequent
fail, action, or node pass to compute its result without backtracking on rules.

## Fail pass

The fail pass generates a function for each rule of the grammar twith a signature of the form:
```
func <Prefix><RuleName>Fail(parser *<Prefix>Parser, start, errPos int) (int, *peg.Fail)
```

The functions of the fail pass assume that the `Parser` has already been used
as the argument of a corresponding accept pass,
and that the accept pass failed to accept.

Each function returns the `*peg.Fail` tree of all attempted rules
that failed to accept the input beginning from `start`,
which failed no earlier than `errPos` bytes into the input.

The description is somewhat advanced.
Suffice it to say, this computes a data structured used by the `peg` package
to compute a parse error string with the `peg.SimpleError` function.
More advanced users can inspect the `*peg.Fail` tree
to create more precise or informative parse errors.

## Action pass

The action pass generates a function for each rule of the grammar twith a signature of the form:
```
func <Prefix><RuleName>Action(parser *<Prefix>Parser, start int) (int, *<RuleType>)
```

The functions of the action pass assume that the `Parser` has already been used
as the argument of a corresponding accept pass,
and that the accept pass accepted the rule at this position.

Each function returns the number of consumed runes
and a pointer to a value of the rule expression's result type.

## Node pass

The node pass generates a function for each rule of the grammar twith a signature of the form:
```
func <Prefix><RuleName>Node(parser *<Prefix>Parser, start int) (int, *peg.Node)
````

The functions of the node pass assume that the `Parser` has already been used
as the argument of a corresponding accept pass,
and that the accept pass accepted the rule at this position.

Each function returns the number of consumed runes
and a *peg.Node that is the root of the syntax tree of the parse.

(Peggy is not an official Google product.)