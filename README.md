# **Execução da Ferramenta WasmManipulator**
A execução da ferramenta WasmManipulator é feita através do executável wmr. Esta suporta um conjunto de opções que permitem providenciar ao utilizador a capacidade de configurar a execução da mesma. A configuração da ferramenta pode ser feita através de variáveis de ambiente ou da passagem de parâmetros ao iniciar a sua execução. No commando representado abaixo encontra-se um exemplo de um comando usado para executar a ferramenta, onde o nome do módulo WASM de entrada é definido numa variável ambiente (WMR\_IN\_MODULE="module.wasm"), e os advices incluídos na execução são definidos no parâmetro de execução (--include=advice\_1,advice\_2).

> WMR\_IN\_MODULE="module.wasm" **./wmr** --include=advice\_1,advice\_2

Na seguinte tabela encontra-se um resumo das configurações existentes na ferramenta. Cada configuração tem um tipo específico e um valor predefinido.

||***Variável Ambiente***|***Parâmetro***|***Tipo***|***Default***|
| - | - | - | - | - |
|***Ficheiro de entrada do módulo***|WMR\_IN\_MODULE|in\_module|*string*|input.wasm|
|***Ficheiro de entrada da transformação***|WMR\_IN\_TRANSFORM|in\_transform|*string*|input.yml|
|***Ficheiro de saída do módulo transformado***|WMR\_OUT\_MODULE|out\_module|*string*|output.wasm|
|***Ficheiro de saída do JS auxiliar***|WMR\_OUT\_JS|out\_js|*string*|output.js|
|***Ficheiro de saída do módulo original***|WMR\_OUT\_MODULE\_ORIG|out\_module\_orig|*string*|*null*|
|***Diretoria com as dependências***|WMR\_DEPENDENCIES\_DIR|dependencies\_dir|*string*|./dependencies/|
|***Diretoria com dados para execução***|WMR\_DATA\_DIR|data\_dir|*string*|./|
|***Ficheiro de logs***|WMR\_LOG\_FILE|log\_file|*string*|*null*|
|***Incluir advices***|WMR\_INCLUDE|include|*string[]*|Todos|
|***Excluir advices***|WMR\_EXCLUDE|exclude|*string[]*|Nenhum|
|***Gerar sempre o ficheiro JS***|WMR\_PRINT\_JS|print\_js|*boolean*|*false*|
|***Gerar sempre o módulo transformado***|WMR\_ALLOW\_EMPTY|allow\_empty|*boolean*|*false*|
|***Gerar todo os logs***|WMR\_VERBOSE|verbose|*boolean*|*false*|
|***Não ordenar advices***|WMR\_IGNORE\_ORDER|ignore\_order|*boolean*|*false*|

<br>

## **Opções de Configuração**
**Ficheiro de entrada do módulo**

Nesta configuração o utilizador deve inserir o ficheiro de entrada que pretende modificar. O ficheiro pode encontrar-se no formato binário (.wasm) ou no formato textual (.wat), e deve conter um módulo WASM válido.

Exemplos:

- ./wmr --data\_dir="$HOME/" --in\_module="data/module.wasm" (o ficheiro encontra-se em $HOME/data/module.wasm)
- WMR\_IN\_MODULE="data/module.wasm" ./wmr (o ficheiro encontra-se em ./data/module.wasm)

**Ficheiro de entrada da transformação**

A configuração do ficheiro de entrada que contém as instruções de transformação do módulo deve assumir o formato YAML (YAML Ain't Markup Language).

Exemplos:

- ./wmr --in\_transform="data/transf.yml" (o ficheiro encontra-se em ./data/transf.yml)
- WMR\_IN\_TRANSFORM="data/transf.yml" ./wmr (o ficheiro encontra-se em ./data/transf.yml)

**Ficheiro de saída do módulo transformado**

Esta configuração representa o ficheiro resultante das transformações. Este ficheiro será constituído por um módulo válido no formato binário.

Exemplos:

- ./wmr --out\_module="result.wasm" (o ficheiro encontra-se em ./result.wasm)
- WMR\_OUT\_MODULE="result.wasm" ./wmr (o ficheiro encontra-se em ./result.wasm)

**Ficheiro de saída do JS auxiliar**

O ficheiro de saída auxiliar ao módulo transformado assume o formato JS (.js). A menos que a configuração “Gerar sempre o ficheiro JS” esteja ativa, este ficheiro nem sempre será criado após a execução da ferramenta. Isto deve-se ao facto do código JS apenas ser necessário caso o utilizador utilize tipos complexos ou *runtime expressions* na transformação do módulo.

Exemplos:

- ./wmr --out\_js="result.js" (o ficheiro encontra-se em ./result.js)
- WMR\_OUT\_JS="result.js" ./wmr (o ficheiro encontra-se em ./result.js)

**Ficheiro de saída do módulo original**

Nesta configuração é indicado se o ficheiro binário inicial no qual sofreu as transformações deverá ser criado. Este ficheiro é criado caso a configuração esteja definida corretamente e o ficheiro de entrada com o módulo a ser transformado seja do tipo textual (.wat).

Exemplos:

- ./wmr --out\_module\_orig="module\_orig.wasm" (o ficheiro encontra-se em ./module\_orig.wasm)
- WMR\_OUT\_MODULE\_ORIG="module\_orig.wasm" ./wmr (o ficheiro encontra-se em ./module\_orig.wasm)

**Diretoria com as dependências**

Consiste no caminho base (absoluto ou relativo) onde se encontram as dependências necessárias para a execução. A partir desse caminho, devem existir os seguintes executáveis:

- ${WMR\_DEPENDENCIES\_DIR}/wabt/wasm2wat
- ${WMR\_DEPENDENCIES\_DIR}/wabt/wat2wasm
- ${WMR\_DEPENDENCIES\_DIR}/minifyjs/bin/minify.js
- ${WMR\_DEPENDENCIES\_DIR}/comby/comby

Por predefinição o caminho para as dependências é "./dependencies", sendo que inicia na diretoria onde foi executada a ferramenta.

Esta configuração pode ser ignorada caso o utilizador possua os executáveis definidos na variável de ambiente "PATH". Esta opção não é recomendada, uma vez que a versão instalada pode causar problemas na execução da ferramenta.

Exemplos:

- ./wmr --dependencies\_dir="$HOME/"
- WMR\_DEPENDENCIES\_DIR="$HOME/" ./wmr

**Diretoria com dados para execução**

Consiste no caminho (absoluto ou relativo) onde se encontram os dados de entrada necessários à ferramenta, e onde os resultados serão armazenados após terminar a execução.

Por predefinição o caminho da diretoria dos dados de entrada é o mesmo que a diretoria onde foi executada a ferramenta.

Exemplos:

- ./wmr --data\_dir="$HOME/"
- WMR\_DATA\_DIR="$HOME/" ./wmr

**Ficheiro de logs**

O ficheiro de logs é uma configuração opcional, sendo que por predefinição o logs são imprimidos para a consola. Caso a configuração esteja definida com um ficheiro válido, os logs são imprimidos nesse ficheiro, e não para a consola. Atenção, ficheiros com o mesmo nome serão completamente substituídos por este novo.

Exemplos:

- ./wmr --log\_file="logs" (o ficheiro encontra-se em ./logs)
- WMR\_LOG\_FILE="logs" ./wmr (o ficheiro encontra-se em ./logs)

**Incluir *advices***

Esta configuração recebe um array com os nomes dos *advices* que devem ser incluídos na transformação. Esta filtragem permite ao utilizador aplicar apenas os *advices* pretendidos, e assim obter resultados diferentes, para o mesmo ficheiro de transformação.

Exemplos:

- ./wmr --include=advice\_1,advice\_2
- WMR\_INCLUDE=advice\_1,advice\_2 ./wmr

**Excluir *advices***

Esta configuração recebe um array com os nomes dos *advices* que devem ser excluídos na transformação. Esta filtragem permite ao utilizador remover *advices* indesejados, e assim obter resultados diferentes, para o mesmo ficheiro de transformação.

Quando definida juntamente com a configuração “Incluir *advices”*, a remoção dos *advices* é feita com base nos que são resultantes dessa configuração.

Exemplos:

- ./wmr --exclude=advice\_1,advice\_2
- WMR\_EXCLUDE=advice\_1,advice\_2 ./wmr

**Gerar sempre o ficheiro JS**

Quando esta configuração está ativa, o ficheiro auxiliar JS é sempre criado após a execução da ferramenta, independentemente se é necessário para a integração do módulo WASM numa aplicação ou não.

Por predefinição esta configuração está inativa, sendo que, caso seja necessário para o resultado da ferramenta, o JS é gerado.

Exemplos:

- ./wmr --print\_js
- WMR\_PRINT\_JS=true ./wmr

**Gerar sempre o módulo transformado**

Ao ativar esta configuração, o módulo resultante da transformação é sempre gerado. Isto significa que a execução dar-se-á mesmo que não sejam encontrados *join-points* no módulo para os *advices* definidos.

Apesar do módulo resultante não possuir qualquer transformação ao código existente, este pode conter código novo inserido pelo utilizador. Este código pode ser inserido através de variáveis globais ou novas funções que interagem entre si ou com os elementos existentes no módulo original (apenas com recurso a índices numéricos).

Exemplos:

- ./wmr --allow\_empty
- WMR\_PRINT\_JS=true ./wmr

**Imprimir todos os logs**

Esta configuração permite que logs de rastreamento sejam imprimidos juntamente com os restantes logs. Com isto, é possível fornecer ao utilizador logs mais detalhados sobre a execução da aplicação. Quando desativa, apenas imprime logs do tipo *info*, ou seja, logs menos detalhados.

Exemplos:

- ./wmr --print\_js
- WMR\_PRINT\_JS=true ./wmr

**Não ordenar advices**

Indica se deve ou não ordenar os advices segundo o campo "Order". Por predefinição os *advices* são ordenados, sendo que caso não seja indicado o valor no campo "Order", o *advice* é colocado no fim da lista de execução.

Exemplos:

- ./wmr --ignore\_order
- WMR\_IGNORE\_ORDER=true ./wmr

**Nota:**

Qualquer caminho inserido será relativo à diretoria com os dados para execução, isto é, o caminho terá por base o caminho definido na configuração “Diretoria com dados para execução”.

---

# **Especificação da Linguagem do WasmManipulator**
A linguagem para transformação de WASM utiliza o YAML para organizar e estruturar as instruções. Com isto, a sua estrutura de campos encontra-se ilustrada no código abaixo. À frente de cada campo existe um breve descrição dos mesmos no formato de um comentário.

```js
{
  Pointcuts: Mapa, // possui a definição de pointcuts globais, isto é, que poderão ser utilizados em qualquer advice definido.
  Aspects: Mapa, // possui os dados de transformação do módulo.
  Context: { // definição e inicialização de elementos no contexto global do módulo.
    Variables: Mapa, // declaração e inicialização das variáveis globais.
    Functions: { // definição das funções a acrescentar ao módulo.
        Variables: Mapa, // declaração e inicialização das variáveis locais.
        Args: Array<{ // define a lista de argumentos recebidos pela função.
          Name: string, // referente ao nome do argumento.
          Type: string, // referente ao tipo do argumento.
        }>,
        Result: string, // tipo do valor retornado pela função.
        Code: string, // código da função. O código deve ter o formato WAT e pode possuir expressões específicas da aplicação.
        Imported: { // declara a função como sendo uma função importada.
          Module: string, // nome do módulo onde a definição da função está inserida.
          Field: string, // nome do campo onde a definição da função está inserida (dentro do módulo).
        },
        Exported: string, // declara a função como sendo uma função exportada. Caso a função já tenha sido marcada como uma função importada, esta instrução é ignorada.
    },
  },
  Advices: { // definição dos advices a utilizar na transformação.
    Pointcut(obrigatório): string, // definição do pointcut para o advice. Pointcuts globais podem ser utilizados aqui.
    Variables: Mapa, // declaração e inicialização das variáveis locais a inserir nas funções a aplicar as alterações.
    Advice: string, // código que substituirá os join-points. O código deve ter o formato WAT e pode possuir expressões específicas da aplicação.
    Order: i32, // ordem que o advice deve executar.
    All (default: false): boolean, // indica se todas as funções são utilizadas na execução do \pointcut\, ou seja, para além das funções no código, também devem ser usadas as funções adicionadas pelo utilizador através da ferramenta.
    Smart(default: false): boolean, // indica se a transformação é inteligente.
  },
  Start: string, // código a ser adicionado à função inicial do módulo. O código deve ter o formato WAT e pode possuir expressões específicas da aplicação.
  Templates: Mapa, // possui os templates que poderão ser utilizados nos pointcuts.
}
```

## **Sintaxe**
Para facilitar a especificação da linguagem vão ser utilizados os seguintes tipos:

- *Object* - representa um objeto YAML, isto é, um elemento de chave-valor. Opcionalmente pode ter um tipo de valor específico, e para isso, é utilizada a sintaxe *Object<T>* (em que *T* é o tipo do valor).
- *Array* - representa um *array* YAML, isto é, uma lista. A lista deve ter um valor específico, e por isso, é sempre representada com a seguinte sintaxe *Array<T>* (em que *T* é o tipo do valor).
- *String* - representa um caracteres ASCII que formam um valor textual.
- *Identifier* - consiste numa *String* composta por caracteres alfanuméricos e o símbolo "\_".
- *Type* - consiste numa *String* que representa os tipos de dados existentes na ferramenta.
- *Variable* - consiste numa *String* utilizada para declarar e inicializar variáveis.
- *Code* - representa o código no formato de uma *String*. Este é composto por código WAT e pode possuir determinadas instruções que se encontram especificadas abaixo.
- *CodeFunction* - é um subtipo de *Code*, utilizado apenas em funções.
- *CodeAdvice* - é um subtipo de *Code*, utilizado apenas em *advices*.
- *Pointcut* - consiste numa expressão *String* que representa a definição de um pointcut.
- *PointcutGlobal* - é um subtipo de *Pointcut*, com um formato específico para ser invocado por outros pointcuts.
- *PointcutAdvice* - é um subtipo de *Pointcut*, utilizado diretamente no *advice* e com a capacidade de recolher informação de contexto do módulo.
- *Template* - consiste numa *String* que servirá de template na busca de código.

No código abaixo encontra-se ilustrada a mesma estrutura especificada no seguinte código mas com estes tipos aplicados aos campos.

```js
{
  Pointcuts: PointcutGlobal,
  Aspects: {
    Context: {
      Variables: Map<Variable>,
      Functions: {
          Variables: Map<Variable>,
          Args: Array<{
            Name: Identifier,
            Type: Type,
          }>,
          Result: Type,
          Code: CodeFunction,
          Imported: {
            Module: String,
            Field: String,
          },
          Exported: String,
      },
    },
    Advices: {
      Pointcut: PointcutAdvice,
      Variables: Map<Variable>,
      Advice: CodeAdvice,
      Order: i32,
      All: boolean,
      Smart: boolean,
    },
  },
  Start: CodeFunction,
  Templates: Map<Template>,
}
```

### **String**
Por definição, uma *string* é considerada um tipo de dados que consiste num array de bytes que armazena uma sequência de elementos usando um determinado tipo de codificação (Team, 2006). Contudo, no caso da ferramenta, *String* é uma derivação deste tipo de dados, cujos elementos são sempre caracteres, ou seja, este *array* é considerado sempre um elemento textual, usando ASCII como tipo de codificação.
### **Identifier**
*Identifier* é um elemento textual do tipo *String*, no entanto, suporta apenas caracteres do tipo alfanumérico e o *underscore*. É utilizado para identificadores tais como nomes de funções e variáveis.
### **Type**
O *Type* não é bem um tipo de dados, mas sim uma enumeração de *Strings* com os tipos de dados disponíveis para os elementos WASM na ferramenta. Estes tipos (Gohman, Lepesme, Qwerty2501, Spencer, & Um, 2021) são os seguintes:

- *i32* - inteiro *32-bits.*
- *i64* - inteiro *64-bits.*
- *f32* - real *32-bits* (IEEE 754-2008).
- *f64* - real *64-bits* (IEEE 754-2008).
- *string* - possui as mesmas características do tipo *String.*
- *map*[*string*|*i32*|*f32*]*Type* - estrutura de dados do tipo mapa, ou seja, uma estrutura semelhante a uma tabela que permite indexar valores através de uma chave.
- []*Type* - estrutura de dados do tipo *array*, isto é, uma estrutura equivalente a uma lista de valores.

### **Variable**
O tipo *Variable* consiste numa expressão do tipo *String* que permite declarar e inicializar uma dada variável. Para isso, esta variável deve sempre ser utilizada como valor num objeto YAML, sendo que a chave consistirá no nome da variável.

A sintaxe para uma variável é a seguinte: `@tipo < = @valor >?`.

Como o termo indica, o "tipo" consiste no tipo da variável a declarar. Este é do tipo *Type* e é obrigatório na expressão. A inicialização com "valor" é opcional, sendo que caso não seja incluído, a variável assume o valor nulo associado ao tipo. A informação relativa ao valor encontra-se descrita na seguinte tabela.

|***Tipo***|***Valor***|***Nulo***|***Exemplo***|
| - | - | - | - |
|***i32***|i64|0|-1|
|***i64***|i64|0|1|
|***f32***|f32|0|1.1|
|***f64***|f64|0|-1.1|
|***string***|string|“”|“example”|
|***map***|array<[key,value]>|[]|[["key\_1", 1],["key\_2", 2]]|
|***array***|array<value>|[]|[1,2]|
  
<br>
  
### **Code**
A base para o código utilizado na linguagem da ferramenta é o WAT. Partindo desta base foram adicionadas as seguintes extensões exclusivas à linguagem da ferramenta:

- *Static Expressions* – são expressões interpretadas estaticamente, logo só têm acesso a contexto extático, como por exemplo, o nome de uma função.
- *Runtime Expressions* – são expressões sensíveis ao contexto e interpretadas em tempo de execução.
- *Runtime References* – são referências para variáveis interpretadas em tempo de execução.

As expressões (*static* e *runtime*) têm acesso ao contexto onde são aplicadas, sendo que este varia de acordo com o ambiente onde são utilizadas. O único contexto que é comum a todas as expressões é o contexto global, ou seja, funções e variáveis globais definidas no ficheiro de transformações. Os dados incluídos no contexto são acedidos através do respetivo identificador, por exemplo, se no ficheiro de transformações fosse declarada uma nova variável global com o nome “variable”, no interior das expressões, é necessário utilizar esse nome para substituir o identificador pelo índice da variável. Estas expressões são interpretadas pela ferramenta e numa fase final transformadas em código WAT.

### **CodeFunction**
*CodeFunction* é um subtipo de *Code* que disponibiliza às expressões o acesso ao contexto da função criada. Desta forma, o utilizador consegue aceder aos argumentos da função, às suas variáveis locais, etc.
### **CodeAdvice**

*CodeAdvice* é também um subtipo de *Code* que disponibiliza às expressões o acesso ao contexto do advice. Com isto, as expressões têm acesso não só aos dados definidos no *advice*, mas também à informação fornecida pelos *join-points* encontrados.

O código presente neste elemento consiste no código que irá substituir o conteúdo a que cada *join-point* está associado. Desta forma, segundo as linguagens orientadas a aspetos, este consiste numa operação "*around*". No entanto, com recurso à *keyword* this que permite a inclusão do código associado, o utilizador é capaz de realizar as operações "*before*" e "*after*" sobre o *join-point* em questão.
### **Pointcut**
Tal como na definição do *pointcut*, este tipo tem como objetivo encontrar um conjunto de *join-*points que coincidam com a expressão definida.

A sintaxe de um *Pointcut* varia de acordo com o tipo, contudo é sempre semelhante, parecendo-se com uma função *lambda* do JS:

(@parâmetro <, @parâmetro>\*) => @expressão.

Os "parâmetros" variam de acordo com o tipo de *Pointcut*, no entanto a "expressão" mantém sempre o mesmo formato independentemente do tipo de *Pointcut Expression*.
### **PointcutGlobal**
O *PointcutGlobal* é usado para definir um *pointcut* com propriedades globais, que pode ser incluído nos *pointcuts* associados a *advices*. Com isto, estes não possuem qualquer acesso ao contexto das funções, sendo que os parâmetros passados para o *lambda* são meras variáveis, desconhecidas pelo *pointcut*, e apenas controladas pelo invocador.

A sintaxe para o parâmetro do pointcut global é:

@tipo? @nome.

O "tipo" consiste no tipo da variável, é opcional, e é do tipo *Type*. Quando definido, é criada uma restrição sobre o tipo do parâmetro, quando este não é, o parâmetro pode assumir qualquer tipo. O "nome" é do tipo *Identifier* e consiste no nome da variável que será utilizado como referência na expressão do *Pointcut*.

### **PointcutAdvice**
O *PointcutAdvice* é também usado para definir um *pointcut*, contudo disponibiliza o acesso aos dados de contexto das funções. Estes dados de contexto são passados como parâmetros, e podem ser utilizados tanto na expressão do *Pointcut* como no código do *advice*.

A expressão deste tipo de *Pointcut* pode invocar *Pointcuts* do tipo *PointcutGlobal*, passando-lhes as variáveis de contexto como argumentos.

A sintaxe para o parâmetro deste pointcut é:

> <@tipo\_variável.>?@tipo\_contexto[@índice] @nome.

Para este tipo de *Pointcut* existem dois tipos de dados, o tipo da variável ("tipo\_variável"), e o tipo de contexto ("tipo\_contexto"). O tipo de variável é do tipo *Type* e refere-se à variável em si, o tipo de contexto é mais semelhante a um metadado, e refere-se ao tipo da variável no contexto de uma função. Este é composto por dois tipos: param (parâmetro) ou local (variável local). O "índice" pode assumir um valor numérico (ordem da variável dentro do seu contexto – semelhante ao espaço de índices, no entanto, há uma separação entre variáveis locais e parâmetros) ou o valor do índice em si (evitar uso em relação ao código original visto que no momento da transformação este pode ser imprevisível. O "nome" é do tipo *Identifier* e consiste no nome da variável, sendo este utilizado como referência na expressão do *Pointcut*.

### **Template**
Por fim, o tipo *Template* consiste numa *String* que servirá de padrão na procura de código. Esta procura é feita com recurso à ferramenta Comby.

Para além de texto, o *Template* é composto por uma extensão parecida com as *static expressions*, no entanto, apesar de possuírem a mesma sintaxe, as expressões no *Template* são muito mais limitadas, tendo acesso apenas ao contexto presente no mesmo. Por esta razão, e por forma a distinguir ambos os tipos, estas vão ser chamadas de *template expressions*.

## **Pointcut Expressions**
*Pointcut expressions* são um tipo expressões utilizadas na definição de um *Pointcut*, onde o utilizador combina um conjunto de funções *pointcut* através de operadores lógicos. Neste capítulo vão ser abordadas as várias funções disponibilizadas para criar uma destas expressões e quais os operadores disponíveis na ferramenta.

Os *pointcuts* disponíveis na ferramenta são os seguintes:

- func - encontra funções com uma determinada definição.
- call - encontra chamadas a funções que coincidam com uma determinada definição.
- args - encontra chamadas a funções que sejam chamadas com determinadas restrições nos argumentos.
- returns - encontra as instruções de retorno de um função.
- template - encontra um conjunto de instruções que coincidam com a definição do *template*.

Cada *pointcut* disponibiliza um conjunto de informação ao contexto do advice. Para aceder aos dados do *pointcut*, basta utilizar a *keyword* da função de *pointcut* dentro das expressões.

O acesso a estes dados deve ser feita de forma cautelosa, uma vez que, quando combinados com os operadores lógicos poderão ficar inconsistentes, uma vez que a expressão poderá provocar que determinados pointcuts tornem-se inválidos (por exemplo, na expressão func || args poderão existir situações em que apenas uma das duas funções de *pointcut* exista no *join-point* resultante).

Uma nota sobre os *join-points* encontrados: quando estes se sobrepõem, é sempre escolhido aquele que se encontra num nível de profundidade superior do código. O outro é ignorado, uma vez que se encontra inserido no primeiro. Por exemplo, na situação em que o *pointcut* call é executado sobre a expressão (call $f0 (call $f1)), apesar de coincidirem as instruções (call $f1) e (call $f0 (call $f1), a que prevalece é a exterior ((call $f0 (call $f1))).

### **Pointcut func**
O *pointcut* func filtra os *join-points* de acordo com a definição da função a que pertencem. Isto é, as instruções presentes no *join-point* devem pertencer a uma função que coincida com a configuração definida pelo utilizador.

Caso a execução do *pointcut* seja feita sobre um ambiente vazio (primeira operação a ser executada), este cria um *join-point* para cada função que coincide com a definição, envolvendo todas as instruções presentes nessa função.

#### **Sintaxe**
A sintaxe para o *pointcut* func é:

> func(@retorno @função(@parâmetros?)<, @scope>?).

Os elementos da sintaxe poderão assumir múltiplos valores. Na seguinte tabela estão descritos estes elementos e a respetiva sintaxe.

**Nome**: retorno<br>
**Descrição**: Tipo de retorno<br>
**Observações**:
- "tipo" é do tipo *Type*
- "ident" é do tipo *Identifier*

|***Sintaxe***|***Significado***|***Exemplo***|
| - | - | - |
|***\****|qualquer tipo de retorno|\*|
|***void***|sem retorno|void|
|***@tipo***|tipo do retorno|i32|
|***%@ident%***|designação do tipo fica armazenado numa variável; qualquer tipo de retorno|%var%|
|***%@ident:void%***|designação do tipo fica armazenado numa variável; sem retorno|%var:void%|
|***%@ident:@tipo%***|designação do tipo fica armazenado numa variável; tipo do retorno|%var:i32%|

<br/>

**Nome**: função<br>
**Descrição**: Identificador da função<br>
**Observações**:
- "nome", "ident", "índice\_nome" são do tipo *Identifier*
- "regex" é do tipo *String*
- "índice\_ordem" é do tipo *i32* (*Type*)


|***Sintaxe***|***Significado***|***Exemplo***|
| - | - | - |
|***\****|qualquer identificador|\*|
|***@ nome***|nome exportado da função|fn\_name|
|***/@regex/***|expressão regular para o nome exportado da função|/\w+/|
|***$@índice\_nome***|índice textual da função|$f1|
|***[@índice\_ordem]***|índice de ordem da função|[1]|
|***%@ident%***||%fn%|
|***%@ident:@nome%***|nome exportado fica armazenado numa variável; nome exportado da função|%fn:fn\_name%|
|***%@ident:/@regex/%***|nome exportado fica armazenado numa variável; expressão regular para o nome exportado da função|%fn:/\w+/%|
|***%@ident:$@índice\_nome%***|índice (textual) fica armazenado numa variável; índice textual da função|%fn:$f1%|
|***%@ident:[@índice\_ordem]%***|índice (ordem) fica armazenado numa variável; índice de ordem da função|%fn:[1]%|

<br/>

**Nome**: parâmetros<br>
**Descrição**: Parâmetros da função<br>
**Observações**:
- A sintaxe para o "parâmetro" está definida abaixo

|***Sintaxe***|***Significado***|***Exemplo***|
| - | - | - |
||sem parâmetros||
|***..***|qualquer configuração para os parâmetros|..|
|***@parâmetro <, @parâmetro>\****|tipo do retorno|i32 %p0%, i64|

<br/>

**Nome**: parâmetro<br>
**Descrição**: Parâmetro da função<br>
**Observações**:
- "tipo" é do tipo *Type*
- "ident" é do tipo *Identifier*

|***Sintaxe***|***Significado***|***Exemplo***|
| - | - | - |
|***\****|qualquer parâmetro na respetiva ordem|\*|
|***@tipo***|parâmetro de um tipo específico|i32|
|***\* %@ident%***|parâmetro de qualquer tipo armazenado numa variável|\* %p0%|
|***@tipo %@ident%***|parâmetro de um tipo específico armazenado numa variável|i32 %p0%|

<br/>

**Nome**: scope<br>
**Descrição**: *Scope* da função no módulo<br>

|***Sintaxe***|***Significado***|***Exemplo***|
| - | - | - |
||a função pode conter qualquer *scope*||
|***imported***|função importada|imported|
|***exported***|função exportada|exported|
|***internal***|função interna (privada, ou seja, nem importada, nem exportada)|internal|

#### **Dados de Contexto**
Na seguinte tabela está representado o modelo de dados (Func) que representa a informação que é acrescentada ao contexto do *advice*. Estes dados estão associados à função que possui as instruções incluídas no *join-point*. Os dados estão contidos no identificador func, que poderá ser invocado nas expressões do código.

|***Nome***|***Tipo***|***Descrição***|
| - | - | - |
|***Index***|string|Nome do índice da função.|
|***Order***|i32|Ordem do índice da função.|
|***Name***|string|Caso exportada, consiste no nome exportado da função. Caso contrário, é igual ao nome do índice.|
|***Params***|Array<string>|Lista com os nomes dos índices dos parâmetros.|
|***ParamTypes***|Array<string>|Lista com os tipos dos parâmetros.|
|***TotalParams***|i32|Número total dos parâmetros.|
|***Locals***|Array<string>|Lista com os nomes dos índices das variáveis locais.|
|***LocalTypes***|Array<string>|Lista com os tipos das variáveis locais.|
|***TotalLocals***|i32|Número total das variáveis locais.|
|***ResultType***|string|Tipo do resultado da função.|
|***Code***|string|Instruções da função em formato textual.|
|***IsImported***|boolean|Se a função é importada.|
|***IsExported***|boolean|Se a função é exportada.|
|***IsStart***|boolean|Se a função é executada inicialmente.|

### **Pointcut call**
O *pointcut* call tem como objetivo encontrar instruções que correspondam a chamadas a funções com uma determinada configuração. Os *join-points* gerados pelo *pointcut* correspondem à instrução na globalidade, incluindo não só a instrução da chamada, mas também as instruções correspondentes aos argumentos passados à função.

#### **Sintaxe**
A sintaxe utilizada para a configuração é igual à sintaxe para o *pointcut* func. Isto deve-se ao facto de que ambos dependem da configuração da função para operar.

Com isto, a sintaxe para o *pointcut* call é:

> call(@retorno @função(@parâmetros?)).

A descrição dos vários elementos da sintaxe encontra-se expressa na tabela da secção com o *pointcut* func.

#### **Dados de Contexto**
À semelhança do *pointcut* func, o *pointcut* call acrescenta dados relacionados com a função a que o *join-point* está associado. Contudo, estes dados tanto existem para a função que fez a chamada, como para a função que foi invocada, e por isso são encapsulados campos diferentes. Para além desses campos, são também incluídos os dados relacionados com os argumentos passados na instrução da chamada. Este modelo de dados encontra-se representado na seguinte tabela. Os dados estão contidos no identificador call, que poderá ser invocado nas expressões do código.

|***Nome***|***Tipo***|***Descrição***|
| - | - | - |
|***Callee***|Func (Tabela *func*)|Dados da função invocada.|
|***Caller***|Func (Tabela *func*)|Dados da função que invocou.|
|***Args***|Array<Arg> (Tabela *arg*)|Lista com a informação dos argumentos.|
|***TotalArgs***|i32|Número total de argumentos.|

O objeto Arg possui a informação relativa ao argumento de uma função. O seu modelo de dados encontra-se representado na seguinte tabela.

|***Nome***|***Tipo***|***Descrição***|
| - | - | - |
|***Type***|string|Tipo do argumento.|
|***Order***|i32|Ordem do argumento na chamada.|
|***Instr***|string|Código WAT do argumento.|

### **Pointcut args**
O *pointcut* args, tal como o *pointcut* call, tem como objetivo encontrar chamadas a funções, no entanto, a pesquisa para este é feita com recurso às variáveis de contexto passadas como parâmetros ao *Pointcut*.

Ao aceitar apenas variáveis de contexto para a pesquisa faz com que os resultados a obter sejam muito específicos, uma vez que a instrução da chamada deve ter obrigatoriamente nos seus argumentos o acesso a essas variáveis (instrução local.get).

#### **Sintaxe**
A sintaxe para o *pointcut* args é:

> args(<@argumento <, @argumento>\*>?).

O *pointcut* aceita qualquer número de argumentos, sendo que cada "argumento" é do tipo *Identifier* e corresponde a uma variável de contexto do *Pointcut*.

#### **Dados de Contexto**
O modelo de dados do *pointcut* args é igual ao do *pointcut* call, e por isso encontra-se representado na tabela da respetiva secção do *pointcut*. Este é também composto pelos dados referentes a ambas as funções (a função invocada e a que invocou a chamada) e pelos dados referentes aos argumentos passados na instrução. Os dados estão contidos no identificador args, que poderá ser invocado nas expressões do código.

### ***Pointcut* returns**
O *pointcut* returns tem como objetivo encontrar todas as instruções de retorno de uma dada função. Este aceita na sua configuração um dado tipo, que permite filtrar os *join-points* por tipo de retorno.

#### **Sintaxe**
A sintaxe para o *pointcut* returns é:

> returns(@tipo).

O "tipo" consiste no tipo de dados esperado no retorno, sendo que também é aceite o valor \* para indicar que os *join-points* não requerem nenhum tipo de retorno em específico.

#### **Dados de Contexto**
O modelo de dados correspondente aos dados de contexto adicionados após a execução do *pointcut* returns está representado na seguinte tabela. Estes dados estão relacionados à instrução de retorno a que o *join-point* está associado. Os dados estão contidos no identificador returns, que poderá ser invocado nas expressões do código.

|***Nome***|***Tipo***|***Descrição***|
| - | - | - |
|***Func***|Func (Tabela *func*)|Dados da função que contém a instrução de retorno.|
|***Type***|string|Tipo da instrução de retorno.|
|***Instr***|string|Código WAT do instrução de retorno.|

### ***Pointcut* template**
Este *pointcut* é utilizado para realizar a pesquisa por padrão na ferramenta. Para isso, deve ser referenciado o respetivo *template* que servirá de padrão durante a pesquisa por *join-points*.

#### **Sintaxe**
A sintaxe para o *pointcut* template é:

> template(<@template <, @validação>).

O "template" indicado no *pointcut* corresponde a uma das chaves presente no ficheiro de transformação, dentro do objeto *Templates*, que está associada ao *template* que servirá de padrão na pesquisa.

A "validação" é do tipo *boolean* (true ou false), e serve para indicar se o *template* está a ser executado apenas como forma de validação ou não. Por predefinição esta configuração está desativada, o que significa que os resultados obtidos apenas contém as instruções que coincidem diretamente com a definição do *template*. Ao ativar a configuração, o *template* servirá apenas como um padrão de validação, onde não é feita a filtragem das instruções, e por isso, qualquer entrada que possua no seu conteúdo o padrão definido no *template* é adicionado aos resultados. Desta forma, caso um *join-point* seja válido para um dado *template*, todas as instruções deste *join-point* se mantêm.

#### **Dados de Contexto**
Ao contrário dos outros *pointcuts*, o identificador adicionado ao contexto do *advice* corresponde à chave do *template* incluído na definição, e não o nome da própria função de *pointcut*. Com isto, são extraídos as várias variáveis definidas no *template* e encapsulados no identificador de contexto (chave do *template*). Depois, o seu acesso e manipulação é realizado através das funções disponibilizadas nas *static expressions*.

### **Operadores Lógicos**
Estes *pointcuts* são combinados com recurso aos seguintes operadores lógicos:

- && - corresponde ao operador lógico "*And*".
- || - corresponde ao operador lógico "*Or*".
- () - usado no agrupamento de operações.

## **Code Expressions**

### **Static Expressions**
As *static expressions*, ou expressões estáticas, permitem a manipulação do código, o acesso aos dados do contexto, e a realização de operações sobre informação conhecida em *compile time* (informação estática ou proveniente do contexto do *advice*).

Este tipo de expressões representam o principal sistema utilizado pela ferramenta para implementar no código um paradigma orientado a aspetos. Isto deve-se ao facto de não só permitirem manipular conteúdo estático, como também as instruções presentes no *join-point*. Estas instruções estão disponíveis através do identificador this. Como resultado existe uma forma flexível de interagir com cada *join-point*, onde é possível reproduzir as operações comuns às linguagens AOP, tais como, inserir "antes" ou "depois", "substituir" as instruções, etc. Para além disso, a ferramenta também permite a transformação destes dados através das funções de transformação.

#### **Sintaxe**
A sintaxe presente nas *static expressions* é a seguinte:

> %@variável<:@método>\*%.

A "variável" refere-se ao identificador da variável existente no contexto do *advice* ou da função onde é incluída. Relativamente ao "método", este consiste numa função de transformação, em que a sua utilização segue um paradigma funcional (Noleto, 2020), ou seja, são encadeados de forma imperativa, formando uma sequência de operações que ao receber o mesmo valor, devolvem sempre o mesmo resultado.

#### **Contexto**
As *static expressions* podem ser utilizadas tanto em funções como em *advices*. Desta forma, o contexto depende do sítio onde é aplicada a expressão.

Os dados presentes no contexto disponibilizado para as expressões incluídas na definição de funções são os seguintes:

- Parâmetros da função.
- Variáveis locais.
- Variáveis globais.
- Funções declaradas no ficheiro de transformação.

Relativamente, ao contexto disponibilizado nas expressões incluídas no código de um *advice*, este é composto pelos seguintes dados:

- O código do *join-point* (identificador this).
- Variáveis disponibilizadas pelos *pointcuts*.
- Parâmetros do *pointcut*.
- Variáveis locais definidas no *advice*.
- Variáveis globais.
- Funções declaradas no ficheiro de transformação.

#### **Tipos de Variáveis**
Nas *static expressions* os dados possuem tipos distintos. Cada um deste tipo possui um conjunto de funções de transformação associado, que por sua vez, poderá ter comportamentos também diferentes. Desta forma, foram criados os seguintes tipos de dados:

- *string* - equivalente ao tipo de dados *String*.
- *string\_slice* - consiste num *array* de dados com o tipo *String*.
- *template\_search* - corresponde ao resultado obtido num dado *template*.
- *object* - consiste num objeto composto. Pode assumir o tipo *array*, *mapa*, *objeto*, *string*, *i32*, *i64*, *f32*, *f64* ou *null*.

Quando estas expressões são convertidas para WAT, o seu valor é automaticamente convertido para o respetivo valor do tipo *string*. Neste caso, é invocada a função de transformação string() sobre o resultado da expressão.

#### **Funções de Transformação**
Cada função de transformação recebe um valor de entrada e devolve o respetivo resultado de acordo com a operação executada. O tipo dos dados de entrada/saída varia de acordo com a função aplicada. Para além disto, a configuração dos parâmetros da função também varia com o tipo de função.

Na seguinte tabela encontram-se representadas todas as funções de transformação disponíveis na ferramenta. Para cada função é apresentada uma pequena descrição, a sua sintaxe, exemplos de utilização e os tipos dos valores de entrada e saída.

|***Função***|***Descrição***|
| - | - |
|***string***|Converte o valor de entrada numa *String*.|
||***Sintaxe*:** string().|
||<p>***Exemplos***:</p><p>1. ["1","2","3"]:string() → "123".</p><p>2. object<{k1:"v1"}>:string() → "{\"k1\":\"v1\"}".</p>|
||<p>***Tipos***:</p><p>- string → string .</p><p>- string\_slice → string.</p><p>- template\_search → string.</p><p>- object → string.</p>|
|***type***|Devolve o tipo do valor do valor de entrada.|
||***Sintaxe***: type().|
||<p>***Exemplos***:</p><p>1. ["1","2","3"]:type() → "string\_slice".</p>|
||<p>***Tipos***:</p><p>- string → string .</p><p>- string\_slice → string.</p><p>- template\_search → string.</p><p>- object → string.</p>|
|***order***|Devolve a ordem do índice associado a uma dada função.|
||***Sintaxe*:** order().|
||<p>***Exemplos***:</p><p>1. "$f1":string() → "1".</p>|
||<p>***Tipos***:</p><p>- string → string .</p><p>- string\_slice → string.</p><p>- template\_search → string.</p><p>- object → string.</p>|
|***map***|Cria um novo valor a partir do valor de entrada, chamando uma função específica em cada elemento presente no respetivo valor de entrada.|
||<p>***Sintaxe*:** map((@entrada <, @índice>?) => @expressão).</p><p>A "entrada" consiste no identificador que referencia cada elemento presente no valor de entrada.</p><p>O “índice” é opcional e corresponde ao índice numérico da iteração.</p><p>A "expressão" consiste na expressão que será interpretada e originará uma entrada no valor de saída (no lugar do elemento de entrada). Esta "expressão" será sempre convertida para *string*.</p>|
||<p>***Exemplos***:</p><p>1. ["1","2","3"]:map((v) => "num " + v) → ["num 1","num 2","num 3"].</p><p>2. object<{k1:"v1",k2:"v2"}>:map((v) => v) → ["v1","v2"].</p><p>3. object<["v1","v2"]>:map((v) => v) → ["v1","v2"].</p>|
||<p>***Tipos***:</p><p>- string → string.</p><p>- string\_slice → string\_slice.</p><p>- template\_search → string\_slice.</p><p>- object → string\_slice.</p>|
|***repeat***|Repete o valor de entrada um dado número de vezes.|
||<p>***Sintaxe*:** repeat(@n).</p><p>"n" é um valor numérico referente ao número de vezes que o valor de entrada vai ser repetido. Uma particularidade da função é que quando a entrada é um *array*, a saída não será um *array* de *arrays*, mas sim um *array* com cada valor repetido “n” vezes.</p>|
||<p>***Exemplos***:</p><p>1. "123":repeat(2) → ["123","123"].</p><p>2. ["1","2","3"]:repeat(2) → ["1","1","2","2","3","3"].</p><p>3. object<["1","2","3"]>:repeat(2) → ["1","1","2","2","3","3"].</p><p>4. object<{k1:"1"}>:repeat(2) → ["{\"k1\":\"1\"}","{\"k1\":\"1\"}].</p>|
||<p>***Tipos***:</p><p>- string → string\_slice.</p><p>- string\_slice → string\_slice.</p><p>- template\_search → string\_slice.</p><p>- object → string\_slice.</p>|
|***join***|Conecta os elementos do valor de entrada utilizando um determinado separador.|
||<p>***Sintaxe*:** join(@separador).</p><p>O "separador" é utilizado para juntar os vários elementos para um resultado do tipo *string*. Este é sempre convertido para *string*.</p>|
||<p>***Exemplos***:</p><p>1. ["1","2"]:join(",") → "1,2".</p><p>2. object<{k1:"1",k2:"2"}>:join(",") → "1,2".</p>|
||<p>***Tipos***:</p><p>- string → string.</p><p>- string\_slice → string.</p><p>- object → string.</p>|
|***split***|Divide a entrada numa *array* de *strings*.|
||<p>***Sintaxe*:** split(@separador).</p><p>O "separador" é utilizado para separar os vários elementos para um resultado do tipo *string*. Este é sempre convertido para *string*.</p>|
||<p>***Exemplos***:</p><p>1. "123":split("") → ["1","2","3"].</p><p>2. "12345":split("2") → ["1","345"].</p>|
||<p>***Tipos***:</p><p>- string → string\_slice.</p>|
|***count***|Devolve o número de elementos do valor de entrada.|
||***Sintaxe*:** count().|
||<p>***Exemplos***:</p><p>1. "321":count()→ "3".</p><p>2. ["4","3","2","1"]:count() → "4".</p><p>3. object<{k1:"1",k2:"2"}>:count() → "2".</p>|
||<p>***Tipos***:</p><p>- string → string .</p><p>- string\_slice → string.</p><p>- template\_search → string.</p><p>- object → string.</p>|
|***contains***|Devolve se o valor de entrada contém um dado valor/chave.|
||<p>***Sintaxe*:** contains(@valor).</p><p>O "valor" é sempre convertido para *string*.</p>|
||<p>***Exemplos***:</p><p>1. "321":contains("2") → "true".</p><p>2. ["4","3","2","1"]:contains("5") → "false".</p><p>3. object<{k1:"1",k2:"2"}>:contains("k2") → "true".</p>|
||<p>***Tipos***:</p><p>- string → string .</p><p>- string\_slice → string.</p><p>- template\_search → string.</p><p>- object → string.</p>|
|***assert***|Interrompe a cadeia de operações caso a condição não seja atendida.|
||<p>***Sintaxe*:** assert((@entrada => @condição).</p><p>A "entrada" consiste no identificador que referencia o valor de entrada.</p><p>A "condição" será sempre transformada sempre num valor booleano.</p>|
||<p>***Exemplos***:</p><p>1. "123":assert((v) => v - 1 == 122) → "123".</p><p>2. "123":assert((v) => v - 1 != 122) → "".</p>|
||<p>***Tipos***:</p><p>- string → string | "".</p><p>- string\_slice → string\_slice | "".</p><p>- template\_search → template\_search | "".</p><p>- object → object | "".</p>|
|***replace***|Substitui conteúdo do valor de entrada, ou parte do mesmo, por um novo valor.|
||<p>***Sintaxe*:** replace(@valor\_anterior, @valor\_novo).</p><p>O "valor\_anterior" consiste no valor que deve ser substituído pelo "valor\_novo". Ambos os parâmetros, "valor\_anterior" e "valor\_novo", serão sempre convertidos para *string*.</p>|
||<p>***Exemplos***:</p><p>1. "123":replace("2","5") → "153".</p><p>2. ["1","2","3"]:replace("1","5") → ["5","2","3"].</p><p>3. object<{k1:"1",k2:"2"}>:replace("k2","3") → object<{k1:"1",k2:"3"}>.</p><p>4. search<{result:"1|2",values:{k1:"1",k2:"2"}}>:remove("k1","3") → search<{result:"3|2",values:{k1:"3",k2:"2"}}>.</p>|
||<p>***Tipos***:</p><p>- string → string.</p><p>- string\_slice → string\_slice.</p><p>- template\_search → template\_search.</p><p>- object → object.</p>|
|***remove***|Remove parte do conteúdo do valor de entrada.|
||<p>***Sintaxe*:** remove(@valor).</p><p>O "valor" corresponde à configuração que será removida do valor de entrada. Este é sempre convertido para *string*.</p>|
||<p>***Exemplos***:</p><p>1. "123":remove("23") → "1".</p><p>2. ["1","2","3"]:remove("2") → ["1","3"].</p><p>3. object<{k1:"1",k2:"2"}>:remove("k2") → object<{k1:"1"}>.</p><p>4. search<{result:"1|2",values:{k1:"1",k2:"2"}}>:remove("k1") → search<{result:"|2",values:{k2:"2"}}>.</p>|
||<p>***Tipos***:</p><p>- string → string.</p><p>- string\_slice → string\_slice.</p><p>- template\_search → template\_search.</p><p>- object → object.</p>|
|***filter***|Filtra elementos do valor de entrada segundo uma determinada condição.|
||<p>***Sintaxe*:** filter((@entrada <, @índice>?) => @expressão).</p><p>A "entrada" corresponde ao elemento do valor de entrada.</p><p>O “índice” é opcional e corresponde ao índice numérico da iteração.</p><p>A "expressão" consiste na expressão que será interpretada e dependendo do resultado, o valor será adicionado (ou não) no valor de saída.</p>|
||<p>***Exemplos***:</p><p>1. "147":filter((v)=>v%2!=0) → "17".</p><p>2. ["1","4","7"]:filter((v)=>v%2==0) → ["4"].</p><p>3. object<{k1:"1",k2:"2"}>:filter((v)=>v%2==0) → ["2"].</p><p>4. search<{result:"1|2",values:{k1:"1",k2:"2"}}>:filter((v)=>v!="2") → "1|".</p>|
||<p>***Tipos***:</p><p>- string → string.</p><p>- string\_slice → string\_slice.</p><p>- template\_search → string.</p><p>- object → string\_slice.</p>|
|***slice***|Altera o conteúdo de um valor de entrada selecionando o intervalo indicado.|
||<p>***Sintaxe*:** slice(@início <, @fim>?).</p><p>O "início" e o "fim" correspondem aos índices do intervalo que corresponderá ao valor de saída. Estes valores devem ser numéricos, sendo que o "fim" é opcional, assumindo o tamanho do valor de entrada por predefinição.</p>|
||<p>***Exemplos***:</p><p>1. "147":slice(1) → "47".</p><p>2. ["1","4","7"]:slice(0,1) → ["1"].</p><p>3. object<{k1:"1",k2:"2"}>:slice(1)  → ["2"].</p>|
||<p>***Tipos***:</p><p>- string → string.</p><p>- string\_slice → string\_slice.</p><p>- template\_search → string.</p><p>- object → string\_slice.</p>|
|***splice***|Altera o conteúdo de um valor de entrada removendo o intervalo indicado.|
||<p>***Sintaxe*:** splice(@início <, @fim>?).</p><p>O "início" e o "fim" correspondem aos índices do intervalo a remover. Estes valores devem ser numéricos, sendo que o "fim" é opcional, assumindo o tamanho do valor de entrada por predefinição.</p>|
||<p>***Exemplos***:</p><p>1. "147":splice(1) → "1".</p><p>2. ["1","4","7"]:splice(0,1) → ["4","7"].</p><p>3. object<{k1:"1",k2:"2"}>:splice(1)  → ["1"].</p>|
||<p>***Tipos***:</p><p>- string → string.</p><p>- string\_slice → string\_slice.</p><p>- template\_search → string.</p><p>- object → string\_slice.</p>|
|***select***|Seleciona um sub-valor do template.|
||<p>***Sintaxe*:** select(@ident) .</p><p>O "ident" corresponde ao identificador do valor no resultado do *template*. Este é sempre convertido para *string*.</p>|
||<p>***Exemplos***:</p><p>1. search<{result:"1|2",values:{k1:"1",k2:"2"}}>:select("k1") → search<{result:"1",values:{}}> .</p>|
||<p>***Tipos***:</p><p>- template\_search → template\_search.</p>|
|***reverse***|Reverte a ordem dos elementos de entrada.|
||***Sintaxe*:** reverse().|
||<p>***Exemplos***:</p><p>1. ["1","2","3"]:reverse() → ["3","2","1"].</p><p>2. "123":reverse() → "321".</p>|
||<p>***Tipos***:</p><p>- string → string.</p><p>- string\_slice → string\_slice.</p><p>- template\_search → string.</p><p>- object → object.</p>|

#### **Palavras Reservadas (*Keywords*)**
As *keywords* reservadas para as *static expressions* são as seguintes:

- this que contém as instruções associadas a um *join-point*.
- func, call, args e returns que representam os dados de contexto fornecidos pelos vários *pointcuts*.
- o caractere ; permite a junção de múltiplas *static expressions* numa só.

**Nota:** A ferramenta possui algumas palavras reservadas, ou *keywords*, que têm uma finalidade específica e por isso não podem, ou não devem ser usadas como identificadores de variáveis e funções. A utilização das *keywords* variam de acordo com o tipo de expressão onde são utilizadas.

#### **Observações**
Além da informação exposta neste capítulo, existem as seguintes observações a fazer relativamente às *static expressions*:

- Para além dos identificadores de contexto, é possível iniciar uma sequência de operações sobre um valor estático do tipo *string* (%""%).
- Suportam modificadores numéricos (cálculos) dentro dos *lambda*.
- As *keywords* para estas expressões não são as mesmas que para as *runtime expressions.*
- O resultado final é sempre convertido para *string*.
- Os comentários inline (iniciados através de ;;) devem possuir uma linha vazia extra para que sejam interpretados como tal. Caso contrário, o código escrito nessa linha é ignorado. Desta forma, o aconselhável é a utilização de blocos de comentários (entre (; e ;)).

### **Runtime Expressions**
O objetivo principal destas expressões consiste na geração de código em *runtime*, ou seja, o código das transformações é sensível ao contexto em que se encontra a executar. Para além disto, também serão utilizadas para permitir a interação com valores cujo tipo é desconhecido pelo WASM (*strings*, mapas e *arrays*).

As *runtime expressions* estão fortemente acopladas com o JS, uma vez que todo o processo de execução vai ser realizado com recurso à função JS eval (MDN Contributors, 2021).

Com a utilização destas expressões não só será possível a utilização de novos tipos, conseguindo assim a implementação de algumas funcionalidades que utilizando WASM "puro" seriam impossíveis (ou quase impossíveis), tais como, *logging* ou *caching*, como também a capacidade de executar expressões complexas com os dados do contexto em tempo de execução.

Apesar da utilização deste tipo de instruções possuir diversos benefícios, também existem algumas desvantagens que podem ser limitadoras para o utilizador. Uma das desvantagens é o facto do utilizador necessitar não só do código WASM gerado, mas também do código JS, sendo que a interação com o programa WASM deve ser feita com recurso a este último ficheiro, e não como o módulo WASM. Outra desvantagem é que o tamanho do ficheiro WASM aumenta consideravelmente de tamanho devido à geração de instruções necessárias à comunicação com o cliente. Por último, pode existir uma diminuição de desempenho no programa, uma vez que o programa não se resume apenas à utilização de instruções WASM.

A utilização de *runtime expressions* possui algumas restrições. Estas restrições estão relacionadas com a instrução onde é invocada. Quando é invocada na "raiz" da função, assume o tipo do retorno da função. Se for invocada dentro das instruções call, local.set/tee e global.set, dependem do tipo do primeiro argumento da instrução, quer seja função no caso da call, quer seja uma variável no caso do local.set/tee e global.set. Por último, dá para incluir estas expressões dentro de instruções WASM onde é possível conhecer o tipo do valor esperado para o respetivo argumento onde a expressão é aplicada (por exemplo, para a instrução i32.add é possível obter os tipos de ambos os parâmetros - *i32*). Todas as restantes instruções não permitem a utilização deste tipo de expressões.

Com isto, é possível definir a seguinte sintaxe para as diversas formas de aplicação destas expressões:

- @índice = valor do índice
- @expressão = código JS + referências
- @referência = nome da variável
- @runtime\_expression = /@expressão/
- @runtime\_reference = #@referência
- @var\_ident = @índice | @runtime\_reference
- @call = (call @índice @runtime\_expression)
- @set\_local = (local.set @var\_ident @runtime\_expression)
- @tee\_local = (local.tee @var\_ident @runtime\_expression)
- @set\_global = (global.set @var\_ident @runtime\_expression)
- As restantes instruções disponíveis não possuem qualquer formato predefinido, sendo que deve ser utilizada a sintaxe "runtime\_expression" no lugar das expressões.

#### **Runtime References**
As *runtime references* têm como objetivo referenciar uma dada variável no código para as operações *runtime*. Estas tanto podem ser usadas para identificar variáveis dentro das *runtime expressions*, como também para referenciar variáveis que serão alteradas através das instruções local.set/tee e global.set ou retornos de uma função.

No que toca ao primeiro caso, as referências vão permitir que a ferramenta consiga identificar quais as variáveis que devem ser substituídas pelo respetivo valor em tempo de compilação, e assim, proceder às respetivas alterações de código. No segundo caso, estas referências devem sempre ser combinadas com *runtime expressions*, uma vez que não só a interpretação dessas expressões é responsável por atribuir a referência correta à *runtime reference*, como também, as variáveis declaradas neste caso não são inspecionadas em tempo de compilação, e desta forma, não serão detetáveis no momento da execução. Como consequência, o valor pode não existir ou encontrar-se num estado obsoleto no momento em que a referência é executada (ver código abaixo). Para contornar o problema, é aconselhado que quando se necessita de uma variável neste tipo de referência, antes exista uma instrução que a utilize numa *runtime expression*. A utilização das referências só é obrigatória quando é feito o acesso a membros de variáveis do tipo mapa ou *array* (por exemplo, array[1] ou mapa["chave"]).

```
(local.set #index /#index/) ;; A utilização duma instrução vai registar a variável index em tempo de compilação.
(local.set #mapa[index] /#valor/) ;; e assim será possível usar na *runtime reference*.
(...)
(local.set #mapa[index] /#valor/) ;; Valor para a variável index pode encontrar-se obsoleto.
(...)
(local.set #mapa[#index] /#valor/) ;; Erro! A referência *#index* está a ser utilizada com índice de outra *runtime reference* sem ser dentro de uma *runtime expression*.
```

#### **Palavras Reservadas (*Keywords*)**
As *keywords* definidas para as *runtime expressions* são todas as *keywords* existentes no JS, e para além disto, a *keyword* return\_, que representa internamente o valor de retorno de uma função com o tipo complexo.

**Nota:** A ferramenta possui algumas palavras reservadas, ou *keywords*, que têm uma finalidade específica e por isso não podem, ou não devem ser usadas como identificadores de variáveis e funções. A utilização das *keywords* variam de acordo com o tipo de expressão onde são utilizadas.

## **Template Expressions**
As *template expressions* são utilizadas para definir o código dos *templates* que poderão ser utilizados na expressão do *pointcut*, para a realização de uma pesquisa por padrão. Durante o processo de pesquisa, as variáveis que vão sendo recolhidas são incluídas no contexto do *advice*, encapsuladas no identificador referente à chave (nome) do *template*. No final da pesquisa, este identificador é convertido num modelo interno do tipo search\_template, onde poderá ser acedido e manipulado com recurso às *static expressions* aplicadas no código de transformação da ferramenta.

Os *templates* estão definidos no objeto *Template* do ficheiro de transformações, e são identificados através do valor da chave onde estão inseridos, isto é, o seu nome.
### **Template Keywords**
Estas expressões são compostas por uma linguagem específica que combina o WAT com uma sintaxe semelhante às *static expressions*, as *template keywords*, mas cujo objetivo é muito distinto. Enquanto que as *static expressions* são interpretadas e convertidas para WAT, as *template keywords* servem como um *placeholder* no padrão que poderá ficar associado a uma dada variável.

As *template keywords* são capazes de combinar *templates* uns com os outros.* Para isso, estas suportam a utilização de funções, que possuem uma sintaxe semelhante às funções de transformação das *static expressions*, e permitem que uma dada variável respeite uma dada restrição segundo outro *template*. Estas restrições não só englobam que as variáveis coincidam (ou não) com o padrão definido no *template* integrado, como também, declaram variáveis que devem ser definidas nesse *template*.

Na seguinte tabela encontram-se representadas as várias funções disponíveis na ferramenta. Para cada função é feita uma breve descrição e apresentada a devida sintaxe. Na sintaxe, o “template" corresponde ao nome do template a integrar, e o "var\_ident" corresponde ao identificador que deve estar definido no template a integrar.

|***Função***|***Descrição***|***Sintaxe***|
| - | - | - |
|***include***|O valor do identificador deve coincidir com o template indicado.|include(@template)|
|***include\_one***|O valor do identificador deve coincidir com pelo menos um dos templates indicados.|include\_one(@template <, @template>\*)|
|***include\_all***|O valor do identificador deve coincidir com todos os templates indicados.|include\_all(@template <, @template>\*)|
|***not\_include***|O valor do identificador não pode coincidir com o template indicado.|not\_include(@template)|
|***not\_include\_one***|O valor do identificador não pode coincidir com nenhum dos templates indicados.|not\_include\_one(@template <, @template>\*)|
|***not\_include\_all***|O valor do identificador não pode coincidir com todos os templates indicados. Ou seja, só não é válido se coincidir com todos os templates.|not\_include\_all(@template <, @template>\*)|
|***define***|O *template* a integrar deve obrigatoriamente definir os identificadores indicados. Desta forma, a função *define* só é permitida quando é procedida pelas funções *include*, *include\_one* e *include\_all*.|define(@var\_ident)|

### **Funcionamento**
A aplicação dos *templates* é feita com recurso ao Comby (Comby, 2021). Na preparação da *query* as *template keywords* são sempre substituídas por um "*Named Match*", permitindo assim que a ferramenta associe uma dada *keyword* à variável correspondente. Os resultados obtidos são interpretados pela ferramenta e armazenados numa estrutura central recursiva, existindo a possibilidade de serem executados mais do que um *template* de acordo com a definição do utilizador. Esta estrutura contém a respetiva iteração com o valor encontrado e os valores das variáveis que compõem essa iteração. A ferramenta utiliza apenas as primeiras iterações encontradas, ou seja, caso sejam encontradas várias correspondências para a mesma *query*, apenas a primeira será utilizada. Esta limitação foi estabelecida com o intuito de simplificar a utilização dos *templates* para o utilizador, uma vez que, após a transformação do código associado à primeira iteração, o código associado às restantes iterações encontrar-se-ia desatualizado, e como consequência, iria ocorrer um erro de execução, ou no pior cenário, o resultado obtido com as transformações seria enganador ou sem sentido para o utilizador. Contudo, é fornecida uma forma de contornar esta limitação que passa pela utilização de vários *advices* com a mesma definição. O único desafio desta abordagem seria conhecer o número de *advices* que é necessário executar, no entanto, o utilizador pode sempre executar a ferramenta até que não existam novas alterações, e assim, garantir que todas as iterações são devidamente transformadas.

## **Modo Inteligente (Smart)**
Este modo inteligente é configurado para cada um dos *advices* declarados no ficheiro de transformações, e define o modo como as transformações irão operar. Caso este modo esteja ativo, a transformação tem em conta o valor de retorno das instruções referentes ao *join-point* em questão, e procede às transformações extras que permitem manter o mesmo valor de retorno.

Neste modo o utilizador pode definir uma instrução target no código do *advice* que será a instrução que servirá de retorno para o código que está a ser modificado. Caso não seja definido nenhum target, a ferramenta faz uma pesquisa pela instrução que anteriormente existia. Se for encontrada, a ferramenta assume a instrução como sendo o target, mas se essa instrução não existir no novo código, não é realizada qualquer transformação inteligente.

Para perceber melhor o conceito, a seguir será apresentado um exemplo conceptual. Neste exemplo, as instruções que se encontram a ser modificadas são ambas as chamadas existentes na instrução de adição. Esta modificação está relacionada com a instrumentação de código, sendo que deve ser adicionada uma função antes e depois de qualquer chamada realizada no código.

* Código WAT original para o modo "inteligente"
```
(i32.add (call $f0) (call $f1))
```

* *Pointcut expression* para transformação no modo "inteligente"
```
() => call(\* \*(..))
```

* Código do *advice* para a transformação no modo "inteligente"
```
(call $before (i32.const %call.Caller.Order%))
(target %this%)
(call $after (i32.const %call.Caller.Order%))
```

* Código WAT resultante sem o modo "inteligente" ativo
```
(i32.add
  (call $before 0) (call $f0) (call $after 0)
  (call $before 1) (call $f1) (call $after 1)
) ;; Instrução incorreta
```

* Código WAT resultante com o modo "inteligente" ativo
```
(call $before 0) (local.set $tmp0 (call $f0)) (call $after 0)
(call $before 1) (local.set $tmp1 (call $f1)) (call $after 1)
(i32.add
  (local.get $tmp0)
  (local.get $tmp1)
)  ;; Instrução correta
```
