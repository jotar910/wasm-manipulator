# **Execution of the WasmManipulator Tool**
The execution of the WasmManipulator tool is carried out through the executable wmr. This supports a set of options that provide the user with the ability to configure its execution. The tool's configuration can be done through environment variables or by passing parameters when starting its execution. The command shown below is an example of a command used to execute the tool, where the name of the input WASM module is defined in an environment variable (WMR_IN_MODULE="module.wasm"), and the list of advices included in the execution are defined in the execution parameter (--include=advice_1,advice_2).

> WMR_IN_MODULE="module.wasm" **./wmr** --include=advice_1,advice_2

The following table summarizes the existing configurations in the tool. Each configuration has a specific type and a default value.

||***Environment Variable***|***Parameter***|***Type***|***Default***|
| - | - | - | - | - |
|***Input module file***|WMR_IN_MODULE|in_module|*string*|input.wasm|
|***Input transformation file***|WMR_IN_TRANSFORM|in_transform|*string*|input.yml|
|***Output file of the transformed module***|WMR_OUT_MODULE|out_module|*string*|output.wasm|
|***Output file of auxiliary JS***|WMR_OUT_JS|out_js|*string*|output.js|
|***Output file of the original module***|WMR_OUT_MODULE_ORIG|out_module_orig|*string*|*null*|
|***Directory with dependencies***|WMR_DEPENDENCIES_DIR|dependencies_dir|*string*|./dependencies/|
|***Directory with data for execution***|WMR_DATA_DIR|data_dir|*string*|./|
|***Log file***|WMR_LOG_FILE|log_file|*string*|*null*|
|***Include advices***|WMR_INCLUDE|include|*string[]*|All|
|***Exclude advices***|WMR_EXCLUDE|exclude|*string[]*|None|
|***Always generate the JS file***|WMR_PRINT_JS|print_js|*boolean*|*false*|
|***Always generate the transformed module***|WMR_ALLOW_EMPTY|allow_empty|*boolean*|*false*|
|***Generate all logs***|WMR_VERBOSE|verbose|*boolean*|*false*|
|***Do not order advices***|WMR_IGNORE_ORDER|ignore_order|*boolean*|*false*|

<br>

## **Configuration Options**
**Input module file**

In this configuration, the user must insert the input file to be modified. The file can be in binary format (.wasm) or textual format (.wat), and must contain a valid WASM module.

Examples:

- ./wmr --data_dir="$HOME/" --in_module="data/module.wasm" (file is located in $HOME/data/module.wasm)
- WMR_IN_MODULE="data/module.wasm" ./wmr (file is located in ./data/module.wasm)

**Input transformation file**

The configuration of the input file that contains the module transformation instructions should assume the YAML format (YAML Ain't Markup Language).

Examples:

- ./wmr --in_transform="data/transf.yml" (file is located in ./data/transf.yml)
- WMR_IN_TRANSFORM="data/transf.yml" ./wmr (file is located in ./data/transf.yml)

**Output file of the transformed module**

This configuration represents the file resulting from the transformations. This file will consist of a valid module in binary format.

Examples:

- ./wmr --out_module="result.wasm" (file is located in ./result.wasm)
- WMR_OUT_MODULE="result.wasm" ./wmr (file is located in ./result.wasm)

**Output file of auxiliary JS**

The output file auxiliary to the transformed module assumes the JS format (.js). Unless the "Always generate the JS file" configuration is active, this file is not always created after the tool's execution. This is because JS code is only necessary if the user uses complex types or runtime expressions in the module transformation.

Examples:

- ./wmr --out_js="result.js" (file is located in ./result.js)
- WMR_OUT_JS="result.js" ./wmr (file is located in ./result.js)

**Output file of the original module**

This configuration indicates whether the initial binary file on which the transformations were applied should be created. This file is created if the configuration is correctly set and the input file with the module to be transformed is of the textual type (.wat).

Examples:

- ./wmr --out_module_orig="module_orig.wasm" (file is located in ./module_orig.wasm)
- WMR_OUT_MODULE_ORIG="module_orig.wasm" ./wmr (file is located in ./module_orig.wasm)

**Directory with dependencies**

Consists of the base path (absolute or relative) where the necessary dependencies for execution are located. From this path, the following executables should exist:

- ${WMR_DEPENDENCIES_DIR}/wabt/wasm2wat
- ${WMR_DEPENDENCIES_DIR}/wabt/wat2wasm
- ${WMR_DEPENDENCIES_DIR}/minifyjs/bin/minify.js
- ${WMR_DEPENDENCIES_DIR}/comby/comby

By default, the path for dependencies is "./dependencies", starting in the directory where the tool was executed.

This configuration can be ignored if the user has the executables defined in the "PATH" environment variable. This option is not recommended, as the installed version may cause problems in the tool's execution.

Examples:

- ./wmr --dependencies_dir="$HOME/"
- WMR_DEPENDENCIES_DIR="$HOME/" ./wmr

**Directory with data for execution**

Consists of the path (absolute or relative) where the input data necessary for the tool are located, and where the results will be stored after finishing the execution.

By default, the path of the input data directory is the same as the directory where the tool was executed.

Examples:

- ./wmr --data_dir="$HOME/"
- WMR_DATA_DIR="$HOME/" ./wmr

**Log file**

The log file is an optional configuration, by default the logs are printed to the console. If the configuration is set with a valid file, the logs are printed in that file, and not to the console. Be aware, files with the same name will be completely replaced by this new one.

Examples:

- ./wmr --log_file="logs" (file is located in ./logs)
- WMR_LOG_FILE="logs" ./wmr (file is located in ./logs)

**Include *advices***

This configuration receives an array with the names of the *advices* that should be included in the transformation. This filtering allows the user to apply only the desired *advices*, and thus obtain different results, for the same transformation file.

Examples:

- ./wmr --include=advice_1,advice_2
- WMR_INCLUDE=advice_1,advice_2 ./wmr

**Exclude *advices***

This configuration receives an array with the names of the *advices* that should be excluded in the transformation. This filtering allows the user to remove unwanted *advices*, and thus obtain different results, for the same transformation file.

When defined together with the “Include *advices”* configuration, the removal of *advices* is done based on those resulting from that configuration.

Examples:

- ./wmr --exclude=advice_1,advice_2
- WMR_EXCLUDE=advice_1,advice_2 ./wmr

**Always generate the JS file**

When this configuration is active, the auxiliary JS file is always created after the tool's execution, regardless of whether it is necessary for the integration of the WASM module in an application or not.

By default, this configuration is inactive, meaning that, if necessary for the result of the tool, the JS is generated.

Examples:

- ./wmr --print_js
- WMR_PRINT_JS=true ./wmr

**Always generate the transformed module**

By activating this configuration, the resulting module from the transformation is always generated. This means that the execution will take place even if no join-points are found in the module for the defined *advices*.

Although the resulting module does not contain any transformation to the existing code, it may contain new code inserted by the user. This code can be inserted through global variables or new functions that interact with each other or with the elements existing in the original module (only with recourse to numerical indices).

Examples:

- ./wmr --allow_empty
- WMR_PRINT_JS=true ./wmr

**Print all logs**

This configuration allows tracing logs to be printed along with the other logs. With this, it is possible to provide the user with more detailed logs about the application's execution. When deactivated, it only prints *info* logs, that is, less detailed logs.

Examples:

- ./wmr --print_js
- WMR_PRINT_JS=true ./wmr

**Do not order advices**

Indicates whether or not to order the advices according to the "Order" field. By default, advices are ordered, and if the value is not indicated in the "Order" field, the advice is placed at the end of the execution list.

Examples:

- ./wmr --ignore_order
- WMR_IGNORE_ORDER=true ./wmr

**Note:**

Any path entered will be relative to the directory with the data for execution, that is, the path will be based on the path defined in the configuration "Directory with data for execution".

---

# **WasmManipulator Language Specification**
The language for WASM transformation uses YAML to organize and structure instructions. With this, its field structure is illustrated in the code below. In front of each field, there is a brief description of them in the format of a comment.

```js
{
  Pointcuts: Map, // has the definition of global pointcuts, i.e., which can be used in any defined advice.
  Aspects: Map, // has the data for module transformation.
  Context: { // definition and initialization of elements in the global context of the module.
    Variables: Map, // declaration and initialization of global variables.
    Functions: { // definition of functions to be added to the module.
        Variables: Map, // declaration and initialization of local variables.
        Args: Array<{ // defines the list of arguments received by the function.
          Name: string, // related to the name of the argument.
          Type: string, // related to the type of the argument.
        }>,
        Result: string, // type of the returned value by the function.
        Code: string, // function code. The code must be in WAT format and may contain specific expressions of the application.
        Imported: { // declares the function as an imported function.
          Module: string, // name of the module where the function definition is inserted.
          Field: string, // name of the field where the function definition is inserted (within the module).
        },
        Exported: string, // declares the function as an exported function. If the function has already been marked as an imported function, this instruction is ignored.
    },
  },
  Advices: { // definition of the advices to use in the transformation.
    Pointcut(required): string, // definition of the pointcut for the advice. Global pointcuts can be used here.
    Variables: Map, // declaration and initialization of local variables to insert in the functions to apply the changes.
    Advice: string, // code that will replace the join-points. The code must be in WAT format and may contain specific expressions of the application.
    Order: i32, // order that the advice must execute.
    All (default: false): boolean, // indicates if all functions are used in the execution of the \pointcut\, that is, in addition to the functions in the code, the functions added by the user through the tool should also be used.
    Smart(default: false): boolean, // indicates if the transformation is intelligent.
  },
  Start: string, // code to be added to the initial function of the module. The code must be in WAT format and may contain specific expressions of the application.
  Templates: Map, // has the templates that can be used in the pointcuts.
}
```

## **Syntax**
To facilitate the specification of the language, the following types will be used:

- *Object* - represents a YAML object, i.e., a key-value element. Optionally, it may have a specific value type, and for this, the syntax *Object<T>* is used (where *T* is the value type).
- *Array* - represents a YAML array, i.e., a list. The list must have a specific value, so it is always represented with the following syntax *Array<T>* (where *T* is the value type).
- *String* - represents ASCII characters that form a textual value.
- *Identifier* - consists of a *String* composed of alphanumeric characters and the "_" symbol.
- *Type* - consists of a *String* representing the data types available in the tool.
- *Variable* - consists of a *String* used to declare and initialize variables.
- *Code* - represents code in the format of a *String*. This is composed of WAT code and may contain certain instructions that are specified below.
- *CodeFunction* - a subtype of *Code*, used only in functions.
- *CodeAdvice* - a subtype of *Code*, used only in *advices*.
- *Pointcut* - consists of a *String* expression representing the definition of a pointcut.
- *PointcutGlobal* - a subtype of *Pointcut*, with a specific format to be invoked by other pointcuts.
- *PointcutAdvice* - a subtype of *Pointcut*, used directly in the *advice* and capable of gathering context information from the module.
- *Template* - consists of a *String* that will serve as a template in code search.

Below is an illustration of the same structure specified in the following code but with these types applied to the fields.

```js
{
  Pointcuts: PointcutGlobal,
  Aspects: {
    Context: {
      Variables: Map<Variable>,
      Functions: {
          Variables: Map<Variable>,
          Args: Array<{
            Name: Identifier,
            Type: Type,
          }>,
          Result: Type,
          Code: CodeFunction,
          Imported: {
            Module: String,
            Field: String,
          },
          Exported: String,
      },
    },
    Advices: {
      Pointcut: PointcutAdvice,
      Variables: Map<Variable>,
      Advice: CodeAdvice,
      Order: i32,
      All: boolean,
      Smart: boolean,
    },
  },
  Start: CodeFunction,
  Templates: Map<Template>,
}
```
### **String**
By definition, a *string* is considered a data type consisting of a byte array that stores a sequence of elements using a certain type of encoding (Team, 2006). However, in the case of the tool, *String* is a derivation of this data type, whose elements are always characters, that is, this *array* is always considered a textual element, using ASCII as the type of encoding.

### **Identifier**
*Identifier* is a textual element of type *String*, however, it only supports alphanumeric characters and the *underscore*. It is used for identifiers such as function names and variables.

### **Type**
*Type* is not exactly a data type, but rather an enumeration of *Strings* with the data types available for WASM elements in the tool. These types (Gohman, Lepesme, Qwerty2501, Spencer, & Um, 2021) are as follows:

- *i32* - 32-bit integer.
- *i64* - 64-bit integer.
- *f32* - 32-bit real (IEEE 754-2008).
- *f64* - 64-bit real (IEEE 754-2008).
- *string* - has the same characteristics as the *String* type.
- *map*[string|i32|f32]*Type* - a map-type data structure, i.e., a structure similar to a table that allows indexing values through a key.
- []*Type* - an array-type data structure, i.e., a structure equivalent to a list of values.

### **Variable**
The *Variable* type consists of a *String*-type expression that allows declaring and initializing a given variable. For this, this variable must always be used as a value in a YAML object, with the key consisting of the variable's name.

The syntax for a variable is as follows: `@type < = @value >?`.

As the term indicates, the "type" consists of the variable's type to declare. This is of the *Type* type and is mandatory in the expression. Initialization with "value" is optional, and if not included, the variable assumes the null value associated with the type. Information related to the value is described in the following table.

|***Type***|***Value***|***Null***|***Example***|
| - | - | - | - |
|***i32***|i64|0|-1|
|***i64***|i64|0|1|
|***f32***|f32|0|1.1|
|***f64***|f64|0|-1.1|
|***string***|string|“”|“example”|
|***map***|array<[key,value]>|[]|[["key_1", 1],["key_2", 2]]|
|***array***|array<value>|[]|[1,2]|

<br>

### **Code**
The basis for the code used in the tool's language is WAT. Based on this base, the following exclusive extensions to the tool's language have been added:

- *Static Expressions* – are expressions interpreted statically, thus only having access to static context, such as the name of a function.
- *Runtime Expressions* – are context-sensitive expressions interpreted at runtime.
- *Runtime References* – are references to variables interpreted at runtime.

Expressions (*static* and *runtime*) have access to the context in which they are applied, and this varies according to the environment in which they are used. The only context that is common to all expressions is the global context, i.e., global functions and variables defined in the transformation file. Data included in the context is accessed through the respective identifier, for example, if a new global variable named “variable” was declared in the transformation file, inside the expressions, this name must be used to replace the identifier with the variable's index. These expressions are interpreted by the tool and in a final stage transformed into WAT code.

### **CodeFunction**
*CodeFunction* is a subtype of *Code* that provides expressions access to the context of the created function. This way, the user can access the function's arguments, its local variables, etc.
### **CodeAdvice**

*CodeAdvice* is also a subtype of *Code* that provides expressions access to the advice context. With this, expressions have access not only to data defined in the *advice*, but also to information provided by the found *join-points*.

The code in this element consists of the code that will replace the content to which each *join-point* is associated. Thus, according to aspect-oriented languages, this consists of an "around" operation. However, with the use of the keyword this that allows the inclusion of associated code, the user can perform the "before" and "after" operations on the respective *join-point*.
### **Pointcut**
As in the definition of *pointcut*, this type aims to find a set of *join-*points that match the defined expression.

The syntax of a *Pointcut* varies according to the type, however, it is always similar, resembling a JS *lambda* function:

(@parameter <, @parameter>*) => @expression.

The "parameters" vary according to the type of *Pointcut*, however, the "expression" always keeps the same format regardless of the type of *Pointcut Expression*.
### **PointcutGlobal**
*PointcutGlobal* is used to define a *pointcut* with global properties, which can be included in *pointcuts* associated with *advices*. With this, they do not have any access to the functions' context, with the parameters passed to the *lambda* being mere variables, unknown to the *pointcut*, and only controlled by the invoker.

The syntax for the global pointcut parameter is:

@type? @name.

The "type" consists of the variable's type, is optional, and is of the *Type* type. When defined, a restriction is created on the type of the parameter, when it is not, the parameter can assume any type. The "name" is of the *Identifier* type and consists of the variable's name that will be used as a reference in the *Pointcut*'s expression.

### **PointcutAdvice**
*PointcutAdvice* is also used to define a *pointcut*, however, it provides access to the functions' context data. This context data is passed as parameters and can be used both in the *Pointcut* expression and in the advice code.

The expression of this type of *Pointcut* can invoke *Pointcuts* of the *PointcutGlobal* type, passing them the context variables as arguments.

The syntax for the parameter of this pointcut is:

> <@variable_type.?@context_type[@index] @name.

For this type of *Pointcut*, there are two types of data, the variable type ("variable_type"), and the context type ("context_type"). The variable type is of the *Type* type and refers to the variable itself, the context type is more similar to metadata, and refers to the type of the variable in the context of a function. This is composed of two types: param (parameter) or local (local variable). The "index" can take a numeric value (order of the variable within its context – similar to the index space, however, there is a separation between local variables and parameters) or the value of the index itself (avoid use concerning the original code since at the time of transformation this can be unpredictable. The "name" is of the *Identifier* type and consists of the variable's name, being used as a reference in the *Pointcut*'s expression.

### **Template**
Finally, the *Template* type consists of a *String* that will serve as a pattern in the code search. This search is done using the Comby tool.

In addition to text, the *Template* is composed of an extension similar to *static expressions*, however, despite having the same syntax, the expressions in the *Template* are much more limited, having access only to the context present in it. For this reason, and to distinguish both types, these will be called *template expressions*.

## **Pointcut Expressions**
*Pointcut expressions* are a type of expressions used in the definition of a *Pointcut*, where the user combines a set of *pointcut* functions through logical operators. In this chapter, the various functions provided to create one of these expressions and the operators available in the tool will be addressed.

The *pointcuts* available in the tool are as follows:

- func - finds functions with a certain definition.
- call - finds calls to functions that match a certain definition.
- args - finds calls to functions that are called with certain restrictions on the arguments.
- returns - finds the return instructions of a function.
- template - finds a set of instructions that match the *template*'s definition.

Each *pointcut* provides a set of information to the advice context. To access the *pointcut* data, just use the *keyword* of the *pointcut* function within the expressions.

Access to this data must be done cautiously, as, when combined with logical operators, they may become inconsistent, as the expression may cause certain pointcuts to become invalid (for example, in the expression func || args there may be situations where only one of the two *pointcut* functions exists in the resulting *join-point*).

A note on the found *join-points*: when they overlap, the one at a higher depth level in the code is always chosen. The other is ignored, as it is embedded in the first. For example, in the situation where the *pointcut* call is executed on the expression (call $f0 (call $f1)), despite the instructions (call $f1) and (call $f0 (call $f1) coinciding, the prevailing one is the outer ((call $f0 (call $f1))).

### **Pointcut func**
The *pointcut* func filters *join-points* according to the definition of the function to which they belong. That is, the instructions present in the *join-point* must belong to a function that matches the configuration defined by the user.

If the execution of the *pointcut* is done in an empty environment (first operation to be executed), it creates a *join-point* for each function that matches the definition, encompassing all the instructions present in that function.

#### **Syntax**
The syntax for the *pointcut* func is:

> func(@return @function(@parameters?)<, @scope>?).

The elements of the syntax can assume multiple values. The following table describes these elements and their respective syntax.

**Name**: return<br>
**Description**: Return type<br>
**Observations**:
- "type" is of type *Type*
- "ident" is of type *Identifier*

|***Syntax***|***Meaning***|***Example***|
| - | - | - |
|***\****|any return type|\*|
|***void***|no return|void|
|***@type***|return type|i32|
|***%@ident%***|type designation stored in a variable; any return type|%var%|
|***%@ident:void%***|type designation stored in a variable; no return|%var:void%|
|***%@ident:@type%***|type designation stored in a variable; return type|%var:i32%|

<br/>

**Name**: function<br>
**Description**: Function identifier<br>
**Observations**:
- "name", "ident", "index_name" are of type *Identifier*
- "regex" is of type *String*
- "index_order" is of type *i32* (*Type*)


|***Syntax***|***Meaning***|***Example***|
| - | - | - |
|***\****|any identifier|\*|
|***@ name***|exported name of the function|fn_name|
|***/@regex/***|regular expression for the exported name of the function|/\w+/|
|***$@index_name***|textual index of the function|$f1|
|***[@index_order]***|order index of the function|[1]|
|***%@ident%***||%fn%|
|***%@ident:@name%***|exported name stored in a variable; exported name of the function|%fn:fn_name%|
|***%@ident:/@regex/%***|exported name stored in a variable; regular expression for the exported name of the function|%fn:/\w+/%|
|***%@ident:$@index_name%***|index (textual) stored in a variable; textual index of the function|%fn:$f1%|
|***%@ident:[@index_order]%***|index (order) stored in a variable; order index of the function|%fn:[1]%|

<br/>

**Name**: parameters<br>
**Description**: Function parameters<br>
**Observations**:
- The syntax for "parameter" is defined below

|***Syntax***|***Meaning***|***Example***|
| - | - | - |
||no parameters||
|***..***|any configuration for the parameters|..|
|***@parameter <, @parameter>\****|return type|i32 %p0%, i64|

<br/>

**Name**: parameter<br>
**Description**: Function parameter<br>
**Observations**:
- "type" is of type *Type*
- "ident" is of type *Identifier*

|***Syntax***|***Meaning***|***Example***|
| - | - | - |
|***\****|any parameter in the respective order|\*|
|***@type***|parameter of a specific type|i32|
|***\* %@ident%***|parameter of any type stored in a variable|\* %p0%|
|***@type %@ident%***|parameter of a specific type stored in a variable|i32 %p0%|

<br/>

**Name**: scope<br>
**Description**: Function's *scope* in the module<br>

|***Syntax***|***Meaning***|***Example***|
| - | - | - |
||the function can have any *scope*||
|***imported***|imported function|imported|
|***exported***|exported function|exported|
|***internal***|internal function (private, i.e., neither imported nor exported)|internal|

#### **Context Data**
The following table represents the data model (Func) that provides information added to the context of the *advice*. These data are associated with the function that contains the instructions included in the *join-point*. The data are contained in the identifier func, which can be invoked in the code expressions.

|***Name***|***Type***|***Description***|
| - | - | - |
|***Index***|string|Name of the function's index.|
|***Order***|i32|Order of the function's index.|
|***Name***|string|If exported, consists of the exported name of the function. Otherwise, it is equal to the index name.|
|***Params***|Array<string>|List of the names of the parameter indices.|
|***ParamTypes***|Array<string>|List of the types of the parameters.|
|***TotalParams***|i32|Total number of parameters.|
|***Locals***|Array<string>|List of the names of the local variable indices.|
|***LocalTypes***|Array<string>|List of the types of the local variables.|
|***TotalLocals***|i32|Total number of local variables.|
|***ResultType***|string|Type of the function's result.|
|***Code***|string|Instructions of the function in textual format.|
|***IsImported***|boolean|Whether the function is imported.|
|***IsExported***|boolean|Whether the function is exported.|
|***IsStart***|boolean|Whether the function is initially executed.|

### **Pointcut call**
The *pointcut* call aims to find instructions that correspond to calls to functions with a certain configuration. The *join-points* generated by the *pointcut* correspond to the instruction in its entirety, including not only the call instruction but also the instructions corresponding to the arguments passed to the function.

#### **Syntax**
The syntax used for the configuration is the same as the syntax for the *pointcut* func. This is because both depend on the function's configuration to operate.

Thus, the syntax for the *pointcut* call is:

> call(@return @function(@parameters?)).

The description of the various elements of the syntax is expressed in the table in the section with the *pointcut* func.

#### **Context Data**
Similar to the *pointcut* func, the *pointcut* call adds data related to the function associated with the *join-point*. However, these data exist for both the function that made the call and the function that was invoked and therefore are encapsulated in different fields. In addition, data related to the arguments passed in the call instruction are also included. This data model is represented in the following table. The data are contained in the identifier call, which can be invoked in the code expressions.

|***Name***|***Type***|***Description***|
| - | - | - |
|***Callee***|Func (Table *func*)|Data of the invoked function.|
|***Caller***|Func (Table *func*)|Data of the function that invoked.|
|***Args***|Array<Arg> (Table *arg*)|List with information about the arguments.|
|***TotalArgs***|i32|Total number of arguments.|

The Arg object contains information related to a function's argument. Its data model is represented in the following table.

|***Name***|***Type***|***Description***|
| - | - | - |
|***Type***|string|Type of the argument.|
|***Order***|i32|Order of the argument in the call.|
|***Instr***|string|WAT code of the argument.|

### **Pointcut args**
The *pointcut* args, like the *pointcut* call, aims to find calls to functions, however, the search for this is done using context variables passed as parameters to the *Pointcut*.

By accepting only context variables for the search makes the results to be obtained very specific, since the call instruction must necessarily have in its arguments access to these variables (local.get instruction).

#### **Syntax**
The syntax for the *pointcut* args is:

> args(<@argument <, @argument>\*>?).

The *pointcut* accepts any number of arguments, each "argument" being of the *Identifier* type and corresponding to a context variable of the *Pointcut*.

#### **Context Data**
The data model of the *pointcut* args is the same as that of the *pointcut* call, and therefore it is represented in the table of the respective section of the *pointcut*. It also includes data related to both functions (the invoked function and the one that made the call) and data related to the arguments passed in the instruction. The data are contained in the args identifier, which can be invoked in the code expressions.

### ***Pointcut* returns**
The *pointcut* returns aims to find all the return instructions of a given function. It accepts a certain type in its configuration, which allows filtering the *join-points* by return type.

#### **Syntax**
The syntax for the *pointcut* returns is:

> returns(@type).

The "type" consists of the expected data type in the return, and the value \* is also accepted to indicate that the *join-points* do not require any specific type of return.

#### **Context Data**
The data model corresponding to the context data added after the execution of the *pointcut* returns is represented in the following table. These data are related to the return instruction that the *join-point* is associated with. The data are contained in the returns identifier, which can be invoked in the code expressions.

|***Name***|***Type***|***Description***|
| - | - | - |
|***Func***|Func (Table *func*)|Data of the function that contains the return instruction.|
|***Type***|string|Type of the return instruction.|
|***Instr***|string|WAT code of the return instruction.|

### ***Pointcut* template**
This *pointcut* is used to perform pattern search in the tool. For this, the respective *template* that will serve as a pattern during the search for *join-points* must be referenced.

#### **Syntax**
The syntax for the *pointcut* template is:

> template(<@template <, @validation>).

The "template" indicated in the *pointcut* corresponds to one of the keys in the transformation file, within the *Templates* object, which is associated with the template that will serve as a pattern in the search.

The "validation" is of the *boolean* type (true or false), and serves to indicate whether the template is being executed just as a form of validation or not. By default, this configuration is deactivated, which means that the results obtained only contain the instructions that directly match the definition of the template. When activating the configuration, the template will only serve as a validation pattern, where no filtering of instructions is done, and therefore, any entry that has in its content the pattern defined in the template is added to the results. Thus, if a *join-point* is valid for a given template, all instructions of this *join-point* remain.

#### **Context Data**
Unlike other *pointcuts*, the identifier added to the context of the *advice* corresponds to the key of the template included in the definition, and not the name of the pointcut function itself. With this, the various variables defined in the template are extracted and encapsulated in the context identifier (template key). Then, their access and manipulation are performed through the functions available in the *static expressions*.

### **Logical Operators**
These *pointcuts* are combined using the following logical operators:

- && - corresponds to the logical operator "*And*".
- || - corresponds to the logical operator "*Or*".
- () - used in the grouping of operations.

## **Code Expressions**

### **Static Expressions**
*Static expressions*, or static expressions, allow the manipulation of code, access to context data, and the performance of operations on information known at *compile time* (static information or from the context of the *advice*).

This type of expressions represents the main system used by the tool to implement an aspect-oriented paradigm in the code. This is because they not only allow manipulation of static content, but also the instructions present in the *join-point*. These instructions are available through the this identifier. As a result, there is a flexible way to interact with each *join-point*, where it is possible to reproduce common operations of AOP languages, such as inserting "before" or "after", "replacing" instructions, etc. In addition, the tool also allows the transformation of these data through transformation functions.

#### **Syntax**
The syntax present in the *static expressions* is as follows:

> %@variable<:@method>\*%.

The "variable" refers to the identifier of the variable existing in the context of the *advice* or the function where it is included. Regarding the "method", it consists of a transformation function, in which its use follows a functional paradigm (Noleto, 2020), that is, they are chained imperatively, forming a sequence of operations that when receiving the same value, always return the same result.

#### **Context**
*Static expressions* can be used both in functions and in *advices*. Thus, the context depends on where the expression is applied.

The data present in the context provided for the expressions included in the definition of functions are as follows:

- Function parameters.
- Local variables.
- Global variables.
- Functions declared in the transformation file.

Regarding the context provided in the expressions included in the code of an *advice*, it is composed of the following data:

- The code of the *join-point* (identifier this).
- Variables provided by the *pointcuts*.
- *Pointcut* parameters.
- Local variables defined in the *advice*.
- Global variables.
- Functions declared in the transformation file.

#### **Variable Types**
In *static expressions*, data have distinct types. Each of these types has a set of associated transformation functions, which in turn, may have different behaviors. Thus, the following data types were created:

- *string* - equivalent to the *String* data type.
- *string_slice* - consists of an *array* of data with the *String* type.
- *template_search* - corresponds to the result obtained in a given template.
- *object* - consists of a composite object. It can be of the type *array*, *map*, *object*, *string*, *i32*, *i64*, *f32*, *f64*, or *null*.

When these expressions are converted to WAT, their value is automatically converted to the respective *string* type. In this case, the string() transformation function is invoked on the result of the expression.

#### **Transformation Functions**
Each transformation function receives an input value and returns the respective result according to the operation performed. The type of the input/output data varies according to the applied function. In addition, the configuration of the function parameters also varies with the type of function.

The following table represents all the transformation functions available in the tool. For each function, a brief description, its syntax, examples of use, and the types of input and output values are presented.

|***Function***|***Description***|
| - | - |
|***string***|Converts the input value into a *String*.|
||***Syntax***: string().|
||<p>***Examples***:</p><p>1. ["1","2","3"]:string() → "123".</p><p>2. object<{k1:"v1"}>:string() → "{\"k1\":\"v1\"}".</p>|
||<p>***Types***:</p><p>- string → string.</p><p>- string\_slice → string.</p><p>- template\_search → string.</p><p>- object → string.</p>|
|***type***|Returns the type of the input value.|
||***Syntax***: type().|
||<p>***Examples***:</p><p>1. ["1","2","3"]:type() → "string\_slice".</p>|
||<p>***Types***:</p><p>- string → string.</p><p>- string\_slice → string.</p><p>- template\_search → string.</p><p>- object → string.</p>|
|***order***|Returns the order of the index associated with a given function.|
||***Syntax***: order().|
||<p>***Examples***:</p><p>1. "$f1":string() → "1".</p>|
||<p>***Types***:</p><p>- string → string.</p><p>- string\_slice → string.</p><p>- template\_search → string.</p><p>- object → string.</p>|
|***map***|Creates a new value from the input value by calling a specific function on each element present in the respective input value.|
||<p>***Syntax***: map((@input <, @index>?) => @expression).</p><p>"input" refers to the identifier that references each element in the input value.</p><p>"index" is optional and corresponds to the numerical index of the iteration.</p><p>"expression" consists of the expression that will be interpreted and will generate an entry in the output value (in place of the input element). This "expression" will always be converted to *string*.</p>|
||<p>***Examples***:</p><p>1. ["1","2","3"]:map((v) => "num " + v) → ["num 1","num 2","num 3"].</p><p>2. object<{k1:"v1",k2:"v2"}>:map((v) => v) → ["v1","v2"].</p><p>3. object<["v1","v2"]>:map((v) => v) → ["v1","v2"].</p>|
||<p>***Types***:</p><p>- string → string.</p><p>- string\_slice → string\_slice.</p><p>- template\_search → string\_slice.</p><p>- object → string\_slice.</p>|
|***repeat***|Repeats the input value a given number of times.|
||<p>***Syntax***: repeat(@n).</p><p>"n" is a numeric value referring to the number of times the input value will be repeated. A peculiarity of the function is that when the input is an *array*, the output will not be an *array* of *arrays*, but an *array* with each value repeated “n” times.</p>|
||<p>***Examples***:</p><p>1. "123":repeat(2) → ["123","123"].</p><p>2. ["1","2","3"]:repeat(2) → ["1","1","2","2","3","3"].</p><p>3. object<["1","2","3"]>:repeat(2) → ["1","1","2","2","3","3"].</p><p>4. object<{k1:"1"}>:repeat(2) → ["{\"k1\":\"1\"}","{\"k1\":\"1\"}"].</p>|
||<p>***Types***:</p><p>- string → string\_slice.</p><p>- string\_slice → string\_slice.</p><p>- template\_search → string\_slice.</p><p>- object → string\_slice.</p>|
|***join***|Connects the elements of the input value using a specific separator.|
||<p>***Syntax***: join(@separator).</p><p>The "separator" is used to join the various elements into a *string* result. It is always converted to *string*.</p>|
||<p>***Examples***:</p><p>1. ["1","2"]:join(",") → "1,2".</p><p>2. object<{k1:"1",k2:"2"}>:join(",") → "1,2".</p>|
||<p>***Types***:</p><p>- string → string.</p><p>- string\_slice → string.</p><p>- object → string.</p>|
|***split***|Splits the input into an *array* of *strings*.|
||<p>***Syntax***: split(@separator).</p><p>The "separator" is used to separate the various elements into a *string* result. It is always converted to *string*.</p>|
||<p>***Examples***:</p><p>1. "123":split("") → ["1","2","3"].</p><p>2. "12345":split("2") → ["1","345"].</p>|
||<p>***Types***:</p><p>- string → string\_slice.</p>|
|***count***|Returns the number of elements in the input value.|
||***Syntax***: count().|
||<p>***Examples***:</p><p>1. "321":count()→ "3".</p><p>2. ["4","3","2","1"]:count() → "4".</p><p>3. object<{k1:"1",k2:"2"}>:count() → "2".</p>|
||<p>***Types***:</p><p>- string → string.</p><p>- string\_slice → string.</p><p>- template\_search → string.</p><p>- object → string.</p>|
|***contains***|Returns whether the input value contains a given value/key.|
||<p>***Syntax***: contains(@value).</p><p>The "value" is always converted to *string*.</p>|
||<p>***Examples***:</p><p>1. "321":contains("2") → "true".</p><p>2. ["4","3","2","1"]:contains("5") → "false".</p><p>3. object<{k1:"1",k2:"2"}>:contains("k2") → "true".</p>|
||<p>***Types***:</p><p>- string → string.</p><p>- string\_slice → string.</p><p>- template\_search → string.</p><p>- object → string.</p>|
|***assert***|Interrupts the chain of operations if the condition is not met.|
||<p>***Syntax***: assert((@input => @condition).</p><p>"input" refers to the identifier that references the input value.</p><p>"condition" will always be transformed into a boolean value.</p>|
||<p>***Examples***:</p><p>1. "123":assert((v) => v - 1 == 122) → "123".</p><p>2. "123":assert((v) => v - 1 != 122) → "".</p>|
||<p>***Types***:</p><p>- string → string | "".</p><p>- string\_slice → string\_slice | "".</p><p>- template\_search → template\_search | "".</p><p>- object → object | "".</p>|
|***replace***|Replaces content of the input value, or part of it, with a new value.|
||<p>***Syntax***: replace(@old_value, @new_value).</p><p>"old_value" refers to the value to be replaced by "new_value". Both parameters, "old_value" and "new_value", will always be converted to *string*.</p>|
||<p>***Examples***:</p><p>1. "123":replace("2","5") → "153".</p><p>2. ["1","2","3"]:replace("1","5") → ["5","2","3"].</p><p>3. object<{k1:"1",k2:"2"}>:replace("k2","3") → object<{k1:"1",k2:"3"}>.</p><p>4. search<{result:"1|2",values:{k1:"1",k2:"2"}}>:remove("k1","3") → search<{result:"3|2",values:{k1:"3",k2:"2"}}>.</p>|
||<p>***Types***:</p><p>- string → string.</p><p>- string\_slice → string\_slice.</p><p>- template\_search → template\_search.</p><p>- object → object.</p>|
|***remove***|Removes part of the content from the input value.|
||<p>***Syntax:*** remove(@value).</p><p>The "value" corresponds to the configuration that will be removed from the input value. It is always converted to *string*.</p>|
||<p>***Examples***:</p><p>1. "123":remove("23") → "1".</p><p>2. ["1","2","3"]:remove("2") → ["1","3"].</p><p>3. object<{k1:"1",k2:"2"}>:remove("k2") → object<{k1:"1"}>.</p><p>4. search<{result:"1|2",values:{k1:"1",k2:"2"}}>:remove("k1") → search<{result:"|2",values:{k2:"2"}}>.</p>|
||<p>***Types***:</p><p>- string → string.</p><p>- string\_slice → string\_slice.</p><p>- template\_search → template\_search.</p><p>- object → object.</p>|
|***filter***|Filters elements of the input value according to a certain condition.|
||<p>***Syntax:*** filter((@input <, @index>?) => @expression).</p><p>The "input" corresponds to the element of the input value.</p><p>The "index" is optional and corresponds to the numerical index of the iteration.</p><p>The "expression" consists of the expression that will be interpreted and depending on the result, the value will be added (or not) to the output value.</p>|
||<p>***Examples***:</p><p>1. "147":filter((v)=>v%2!=0) → "17".</p><p>2. ["1","4","7"]:filter((v)=>v%2==0) → ["4"].</p><p>3. object<{k1:"1",k2:"2"}>:filter((v)=>v%2==0) → ["2"].</p><p>4. search<{result:"1|2",values:{k1:"1",k2:"2"}}>:filter((v)=>v!="2") → "1|".</p>|
||<p>***Types***:</p><p>- string → string.</p><p>- string\_slice → string\_slice.</p><p>- template\_search → string.</p><p>- object → string\_slice.</p>|
|***slice***|Alters the content of an input value by selecting the indicated range.|
||<p>***Syntax:*** slice(@start <, @end>?).</p><p>The "start" and "end" correspond to the indices of the range that will correspond to the output value. These values must be numeric, with the "end" being optional, assuming the size of the input value by default.</p>|
||<p>***Examples***:</p><p>1. "147":slice(1) → "47".</p><p>2. ["1","4","7"]:slice(0,1) → ["1"].</p><p>3. object<{k1:"1",k2:"2"}>:slice(1)  → ["2"].</p>|
||<p>***Types***:</p><p>- string → string.</p><p>- string\_slice → string\_slice.</p><p>- template\_search → string.</p><p>- object → string\_slice.</p>|
|***splice***|Alters the content of an input value by removing the indicated range.|
||<p>***Syntax:*** splice(@start <, @end>?).</p><p>The "start" and "end" correspond to the indices of the range to be removed. These values must be numeric, with the "end" being optional, assuming the size of the input value by default.</p>|
||<p>***Examples***:</p><p>1. "147":splice(1) → "1".</p><p>2. ["1","4","7"]:splice(0,1) → ["4","7"].</p><p>3. object<{k1:"1",k2:"2"}>:splice(1)  → ["1"].</p>|
||<p>***Types***:</p><p>- string → string.</p><p>- string\_slice → string\_slice.</p><p>- template\_search → string.</p><p>- object → string\_slice.</p>|
|***select***|Selects a sub-value from the template.|
||<p>***Syntax:*** select(@ident) .</p><p>The "ident" corresponds to the identifier of the value in the template result. It is always converted to *string*.</p>|
||<p>***Examples***:</p><p>1. search<{result:"1|2",values:{k1:"1",k2:"2"}}>:select("k1") → search<{result:"1",values:{}}> .</p>|
||<p>***Types***:</p><p>- template\_search → template\_search.</p>|
|***reverse***|Reverses the order of the input elements.|
||***Syntax:*** reverse().|
||<p>***Examples***:</p><p>1. ["1","2","3"]:reverse() → ["3","2","1"].</p><p>2. "123":reverse() → "321".</p>|
||<p>***Types***:</p><p>- string → string.</p><p>- string\_slice → string\_slice.</p><p>- template\_search → string.</p><p>- object → object.</p>|

#### **Reserved Words (*Keywords*)**
The reserved *keywords* for *static expressions* are as follows:

- this, which contains the instructions associated with a *join-point*.
- func, call, args, and returns, which represent the context data provided by various *pointcuts*.
- the character ; allows the joining of multiple *static expressions* into one.

**Note:** The tool has some reserved words, or *keywords*, that have a specific purpose and therefore should not, or must not be used as variable and function identifiers. The use of *keywords* varies according to the type of expression in which they are used.

#### **Observations**
In addition to the information exposed in this chapter, the following observations should be made regarding *static expressions*:

- Besides context identifiers, it is possible to start a sequence of operations on a static value of type *string* (%""%).
- They support numeric modifiers (calculations) within *lambdas*.
- The *keywords* for these expressions are not the same as for the *runtime expressions.*
- The final result is always converted to *string*.
- Inline comments (initiated by ;;) must have an extra blank line to be interpreted as such. Otherwise, the code written on that line is ignored. Therefore, it is advisable to use comment blocks (between (; and ;)).

### **Runtime Expressions**
The main goal of these expressions is to generate code at *runtime*, meaning the transformation code is context-sensitive in its execution. Additionally, they will also be used to interact with values whose type is unknown to WASM (*strings*, maps, and *arrays*).

*Runtime expressions* are closely coupled with JS, as the entire execution process will be carried out using the JS eval function (MDN Contributors, 2021).

The use of these expressions not only allows for the implementation of new types, thereby enabling the implementation of some functionalities that would be impossible (or almost impossible) with pure WASM, such as *logging* or *caching*, but also the ability to execute complex expressions with context data at runtime.

Despite the benefits of using this type of instruction, some disadvantages can be limiting for the user. One disadvantage is the need for not only the generated WASM code but also the JS code, with the interaction with the WASM program having to be done using this latter file, and not as the WASM module. Another disadvantage is that the size of the WASM file increases considerably due to the generation of instructions necessary for communication with the client. Lastly, there may be a decrease in program performance since the program is not limited to using WASM instructions.

The use of *runtime expressions* has some restrictions. These restrictions are related to the instruction where it is invoked. When invoked at the "root" of the function, it assumes the type of the function's return. If invoked within the call, local.set/tee, and global.set instructions, it depends on the type of the first argument of the instruction, whether it be a function in the case of call, or a variable in the case of local.set/tee and global.set. Lastly, these expressions can be included within WASM instructions where it is possible to know the expected value type for the respective argument where the expression is applied (for example, for the i32.add instruction, it is possible to obtain the types of both parameters - *i32*). All other instructions do not allow the use of this type of expressions.

With this, the following syntax can be defined for the various ways these expressions can be applied:

- @index = index value
- @expression = JS code + references
- @reference = variable name
- @runtime_expression = /@expression/
- @runtime_reference = #@reference
- @var_ident = @index | @runtime_reference
- @call = (call @index @runtime_expression)
- @set_local = (local.set @var_ident @runtime_expression)
- @tee_local = (local.tee @var_ident @runtime_expression)
- @set_global = (global.set @var_ident @runtime_expression)
- The remaining available instructions do not have any predefined format, and the "runtime_expression" syntax should be used instead of expressions.

#### **Runtime References**
*Runtime references* aim to reference a given variable in the code for *runtime* operations. They can be used both to identify variables within *runtime expressions* and to reference variables that will be altered through local.set/tee and global.set instructions or function returns.

In the first case, the references will allow the tool to identify which variables should be replaced by their respective value at compile time, and thus proceed with the respective code changes. In the second case, these references must always be combined with *runtime expressions*, as not only the interpretation of these expressions is responsible for assigning the correct reference to the *runtime reference*, but also, the variables declared in this case are not inspected at compile time, and therefore, will not be detectable at the time of execution. As a consequence, the value may not exist or be in an obsolete state when the reference is executed (see code below). To circumvent this problem, it is advisable that when a variable is needed in this type of reference, there should first be an instruction that uses it in a *runtime expression*. The use of references is only mandatory when accessing members of map or *array* type variables (for example, array[1] or map["key"]).

```
(local.set #index /#index/) ;; The use of an instruction will register the variable index at compile time.
(local.set #map [#index] /#value/) ;; and thus it will be possible to use in the runtime reference.
(...)
(local.set #map [#index] /#value/) ;; Value for the variable index may be outdated.
(...)
(local.set #map [#index] /#value/) ;; Error! The reference #index is being used as an index of another runtime reference without being inside a runtime expression.
```

#### **Reserved Words (*Keywords*)**
The *keywords* defined for the *runtime expressions* are all the *keywords* existing in JS, and in addition, the *keyword* return_, which internally represents the return value of a function with a complex type.

**Note:** The tool has some reserved words, or *keywords*, that have a specific purpose and therefore should not, or must not be used as variable and function identifiers. The use of *keywords* varies according to the type of expression where they are used.

## **Template Expressions**
*Template expressions* are used to define the code of the *templates* that can be used in the *pointcut* expression, for pattern search. During the search process, the variables that are collected are included in the context of the *advice*, encapsulated in the identifier corresponding to the key (name) of the *template*. At the end of the search, this identifier is converted into an internal model of the type search_template, where it can be accessed and manipulated using the *static expressions* applied in the tool's transformation code.

The *templates* are defined in the Template object of the transformation file and are identified through the value of the key where they are inserted, that is, their name.
### **Template Keywords**
These expressions are composed of a specific language that combines WAT with syntax similar to *static expressions*, the *template keywords*, but whose purpose is very different. While *static expressions* are interpreted and converted to WAT, *template keywords* serve as a placeholder in the pattern that may be associated with a given variable.

*Template keywords* are capable of combining *templates* with each other.* For this, they support the use of functions, which have a syntax similar to the transformation functions of the *static expressions*, and allow a given variable to respect a given restriction according to another *template*. These restrictions not only include that the variables match (or not) the pattern defined in the integrated *template*, but also declare variables that must be defined in that *template*.

The following table shows the various functions available in the tool. For each function, a brief description is given and the proper syntax is presented. In the syntax, the "template" corresponds to the name of the template to be integrated, and the "var_ident" corresponds to the identifier that must be defined in the template to be integrated.

|***Function***|***Description***|***Syntax***|
| - | - | - |
|***include***|The value of the identifier must match the indicated template.|include(@template)|
|***include_one***|The value of the identifier must match at least one of the indicated templates.|include_one(@template <, @template>*)|
|***include_all***|The value of the identifier must match all the indicated templates.|include_all(@template <, @template>*)|
|***not_include***|The value of the identifier cannot match the indicated template.|not_include(@template)|
|***not_include_one***|The value of the identifier cannot match any of the indicated templates.|not_include_one(@template <, @template>*)|
|***not_include_all***|The value of the identifier cannot match all the indicated templates. That is, it is only invalid if it matches all the templates.|not_include_all(@template <, @template>*)|
|***define***|The *template* to be integrated must necessarily define the indicated identifiers. Therefore, the *define* function is only allowed when preceded by the *include*, *include_one*, and *include_all* functions.|define(@var_ident)|

### **Operation**
The application of *templates* is done using Comby (Comby, 2021). In preparing the *query*, the *template keywords* are always replaced by a "*Named Match*," allowing the tool to associate a given *keyword* with the corresponding variable. The results obtained are interpreted by the tool and stored in a central recursive structure, with the possibility of executing more than one *template* according to the user's definition. This structure contains the respective iteration with the found value and the values of the variables that make up this iteration. The tool only uses the first iterations found, meaning if multiple matches are found for the same *query*, only the first will be used. This limitation was established to simplify the use of *templates* for the user, as after transforming the code associated with the first iteration, the code associated with the remaining iterations would be outdated, and as a consequence, an execution error might occur, or in the worst scenario, the result obtained with the transformations would be misleading or meaningless to the user. However, a way to circumvent this limitation is provided, which involves using several *advices* with the same definition. The only challenge of this approach would be knowing the number of *advices* that need to be executed, but the user can always run the tool until no new changes are found, thus ensuring that all iterations are properly transformed.

## **Smart Mode**
This smart mode is configured for each of the *advices* declared in the transformation file and defines how the transformations will operate. If this mode is active, the transformation takes into account the return value of the instructions related to the *join-point* in question, and proceeds with extra transformations that maintain the same return value.

In this mode, the user can define a target instruction in the *advice* code, which will be the instruction that serves as the return for the code being modified. If no target is defined, the tool searches for the instruction that previously existed. If found, the tool assumes the instruction as the target, but if this instruction does not exist in the new code, no intelligent transformation is performed.

To better understand the concept, a conceptual example will be presented next. In this example, the instructions being modified are both calls existing in the addition instruction. This modification is related to code instrumentation, where a function must be added before and after any call made in the code.

* Original WAT code for "smart" mode
```
(i32.add (call $f0) (call $f1))
```

* *Pointcut expression* for transformation in "smart" mode
```
() => call(* *(..))
```

* *Advice* code for the transformation in "smart" mode
```
(call $before (i32.const %call.Caller.Order%))
(target %this%)
(call $after (i32.const %call.Caller.Order%))
```

* Resulting WAT code without "smart" mode active
```
(i32.add
(call $before 0) (call $f0) (call $after 0)
(call $before 1) (call $f1) (call $after 1)
) ;; Incorrect Instruction
```

* Resulting WAT code with "smart" mode active
```
(call $before 0) (local.set $tmp0 (call $f0)) (call $after 0)
(call $before 1) (local.set $tmp1 (call $f1)) (call $after 1)
(i32.add
(local.get $tmp0)
(local.get $tmp1)
) ;; Correct Instruction
```
