# mtemplate

## Package mtemplate implements data-driven templates for generating textual output such as HTML.

Templates are executed by applying them to a data structure.
Annotations in the template refer to elements of the data
structure (typically a field of a struct or a key in a map)
to control execution and derive values to be displayed.
The template walks the structure as it executes and the
"cursor" @ represents the value at the current location
in the structure.

Data items may be values or pointers; the interface hides the
indirection.

In the following, _field_ is one of several things, according to the data.

- The name of a field of a struct (result = data.field),
- The value stored in a map under that key (result = data[field]), or
- The result of invoking a niladic single-valued method with that name
    (result = data.field())

Major constructs ({} are the default delimiters for template actions;
[] are the notation in this comment for optional elements):

    {# comment }

A one-line comment.

    {.section field} XXX [ {.or} YYY ] {.end}

Set @ to the value of the field.  It may be an explicit @
to stay at the same point in the data. If the field is nil
or empty, execute YYY; otherwise execute XXX.

    {.repeated section field} XXX [ {.alternates with} ZZZ ] [ {.or} YYY ] {.end}

Like *.section*, but field must be an array or slice.  XXX
is executed for each element.  If the array is nil or empty,
YYY is executed instead.  If the {.alternates with} marker
is present, ZZZ is executed between iterations of XXX.

    {! field}
    {! field1 field2 ...}
    {! field|formatter}
    {! field1 field2...|formatter}
    {! field|formatter1|formatter2}

Insert the value of the fields into the output. Each field is
first looked for in the cursor, as in .section and .repeated.
If it is not found, the search continues in outer sections
until the top level is reached.

If the field value is a pointer, leading asterisks indicate
that the value to be inserted should be evaluated through the
pointer.  For example, if x.p is of type *int, {x.p} will
insert the value of the pointer but {*x.p} will insert the
value of the underlying integer.  If the value is nil or not a
pointer, asterisks have no effect.

If a formatter is specified, it must be named in the formatter
map passed to the template set up routines or in the default
set ("html","str","") and is used to process the data for
output.  The formatter function has signature
    func(wr io.Writer, formatter string, data ...interface{})
where wr is the destination for output, data holds the field
values at the instantiation, and formatter is its name at
the invocation site.  The default formatter just concatenates
the string representations of the fields.

Multiple formatters separated by the pipeline character | are
executed sequentially, with each formatter receiving the bytes
emitted by the one to its left.

The delimiter strings get their default value, "{" and "}", from
JSON-template.  They may be set to any non-empty, space-free
string using the SetDelims method.  Their value can be printed
in the output using {.meta-left} and {.meta-right}.


Django-style template inheritance can be achieved by using the
*.parent* and *.child* directives.

A template that expects to be rendered inside of another template
declares which template with the *.parent* directive, and a template
that will have other templates rendered inside of it uses the
*.child* directive to indicate where the included template appears.

For example, if I have a template called detail.html:

    {.parent masterpage.html}
    <h3>Child template.</h3>

and a template called masterpage.html:

    <h2>Parent Page</h2>
    {.child}
    <h2>/Parent Page</h2>

rendering detail.html like this:

    theTemplate := mtemplate.MustParseFile("detail.html", nil)
    outBuffer := new(bytes.Buffer)
    theTemplate.Execute(outBuffer, nil)

will fill outBuffer with:

    <h2>Parent Page</h2>
    <h3>Child template.</h3>
    <h2>/Parent Page</h2>

Templates can be nested in this way as deeply as the developer
needs.