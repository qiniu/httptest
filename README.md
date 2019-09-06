qiniupkg.com/httptest.v1
=============

[![Build Status](https://travis-ci.org/qiniu/httptest.v1.svg?branch=develop)](https://travis-ci.org/qiniu/httptest.v1) [![GoDoc](https://godoc.org/qiniupkg.com/httptest.v1?status.svg)](https://godoc.org/qiniupkg.com/httptest.v1)

[![Qiniu Logo](http://open.qiniudn.com/logo.png)](http://www.qiniu.com/)

# 下载

```
go get -u qiniupkg.com/httptest.v1
```

# 概述

一些背景资料：

* 演讲稿：[七牛如何做HTTP服务测试？](http://open.qiniudn.com/qiniutest.pdf)
* 文字整理稿：[七牛如何做HTTP服务测试？](http://blog.qiniu.com/archives/2541)

这是一套 HTTP 服务测试脚本框架及实用程序。我们定义了一个测试脚本的 DSL 语言。大体看起来是这样的：

```bash
#为了让一套代码同时可以测试 Stage 环境和 Product 环境，我们推荐将 Host、AK/SK 作为环境变量传入
#同时也避免了 AK/SK 这样敏感内容进入代码库
match $(env) `envdecode QiniuTestEnv`
auth qboxtest `qbox $(env.AK) $(env.SK)`
host rs.qiniu.com $(env.RSHost)

post http://rs.qiniu.com/delete/`base64 Bucket:Key`
auth qboxtest
#发起请求，并开始检查结果
ret  200

post http://rs.qiniu.com/batch
auth qboxtest
match $(ekey1) |base64 Bucket:Key|
match $(ekey2) |base64 Bucket2:Key2|
json '{
	"op": ["/delete/$(ekey1)", "/delete/$(ekey2)"]
}'
#发起请求，并开始检查结果
ret  200

post http://rs.qiniu.com/batch
auth qboxtest
form op=/delete/`base64 Bucket:Key`&op=/delete/`base64 Bucket:NotExistKey`
#发起请求，并开始检查结果
ret  298
json '[{"code": $(code1)}, {"code": 612}]'
match $(code1) 200
```

## 命令详解

* 参见：[qiniutest](https://github.com/qiniu/qiniutest)


# 文法

整体是以命令行文法为基础。一条指令由命令及命令参数构成。命令及命令参数之间以空白字符（空格或TAB）分隔。如果某个参数中包含空格或其他特殊字符，则可以：

* 用 \ 转义。比如 '\ ' 表示 ' '(空格)，'\t' 表示 TAB 字符，等等。
* 用 '...' 或 "..." 包含。两者都允许出现 $(var) 或 ${var} 形式表示的变量（这一点和 linux shell 有很大不同，详细见后文 “智能变量” 一节）。他们的区别在于：'...' 中不支持用 \ 转义，也不支持子命令（见后文 “子命令” 一节），出现任何内容都当作普通字符对待。所以 `'\t|abc|'` 用 "..." 来表达必须用 `"\\t\|abc\|"`。

# 类型系统

在 linux shell 的命令行中，所有的输入输出都是字符串，基本上没有类型系统可言。这一点我们和 linux shell 有很大的不同。我们的脚本有完备的类型系统。

我们支持并仅支持 json 文法所支持的所有类型。基础类型包括：number (在 Go 语言中是 float64）、bool、string。复合类型包括 array（在 Go 语言里面是 slice，不是数组）和 dictionary/object (在 Go 语言中是 map/interface{})。特别需要注意的是，我们的类型系统里面没有 int 类型。null 不是空 array，也不是空 dictionary，而是空 object。

由于我们采用了命令行文法，所以表达类型和常规文法有一定的差异。比如 200、'200'、"200" 都表示同一个东西：number 类型的 200。表达字符串 "200" 必须用 '"200"' 或者 "\"200\""。

上面样例中的

```
'[{"code": $(code1)}, {"code": 612}]'
```

是一个 array，如果我们用紧凑文法写，避免任何的空白字符，可以写成这样：

```
[{"code":$(code1)},{"code":612}]
```

当然我们建议表达复合类型的时候，尽量还是用 '...' 来写，以保证可阅读性。

另外，考虑到 json 只有如下这些语法单元：

* array/dictionary/object/string: `[...]`, `{...}`, `null`, `"..."`
* bool: `true`, `false`
* number: `0..9`, `-`
* var: `$(...)` // 我们扩展的语法

我们可以增加一条规则：

* 所有 a..z 或 A..Z 开头的非 `true`, `false`, `null` 文本，被认为是合法的 json string。

也就是说，以下这段文本：

```
http://rs.qiniu.com/batch
```

等价于：

```
'"http://rs.qiniu.com/batch"'
```

另外，对于那些明确接受 string 参数的指令，也可以省略 '"..."' 这样的外衣。


# 智能变量

和 linux shell 类似，我们也支持 $(var) 或 ${var} 格式的变量。但是，$(var) 并不像 linux shell 那样，在命令行词法分析阶段就被处理掉了，它是本 DSL 代表变元的语法成分，和 "..." 是常字符串的语法成分类似。另外，由于我们存在类型系统，所以 $(var) 表达的不是一段文本，而是一个可能是任意类型的 object。这带来这样一些差异：

* 支持 dictionary/object、array 的 member 成员获取操作。比如对 dictionary 可以做 $(a.b.c) 形式的 member 访问。对于 array，理论上应该支持 $(a[1]) 这种形式，不过目前我们用的是 $(a.1)。表达 `a[2].b[3].c` 可以用 $(a.2.b.3.c) 表示。
* 变量智能 marshal。在不同的场景下，变量的 marshal 结果会有差异。所以变量 marshal 需要上下文，而不是简单的字符串替换。比如 http://rs.qiniu.com/delete/$(ekey) 和 {"delete":$(ekey)} 这两个地方，$(ekey) 的 marshal 结果有很大的差别。除了出现在 json 里面的 $(ekey) 需要用 "..." 括起来，而 url 中不需要这样的显著差别外，对于特殊字符的 escape 转义方法也完全不同（但是这个细节经常容易被忽略）。

# 匹配(match)

这几乎是这套 DSL 中最核心的概念。作为一门语言，有变量，自然会有赋值的概念。在这里的确有实现赋值的能力，但它不叫赋值，而是叫匹配。先看例子：

```bash
match $(a.b) 1
match $(a.c) '"hello"'
```

这个例子的结果是，得到了一个变量 a，其值为 {"b": 1, "c": "hello"}。

到现在为止，你看到的 match 像赋值的一面。但是你不能对已经绑定了特定值的变量再次赋不同的值：

```bash
match $(a.b) 1
match $(a.b) 1		#可以匹配，因为$(a.b)的值的确为1
match $(a.b) 2		#失败，1和2不相等
```

match 语句可以很复杂，如：

```bash
match '{"c": {"d": $(d)}}' '{"c": {"d": "hello", "e": "world"}, "f": 1}'
```

一般地，match 命令的文法为：

```bash
match <ExpectedObject> <SourceObject>
```

其中 `<SourceObject>` 中不能出现未绑定的变量。`<ExpectedObject>` 中则允许存在未绑定的变量。`<ExpectedObject>` 和 `<SourceObject>` 不必完全一致，但是 `<ExpectedObject>` 中出现的，在 `<SourceObject>` 中也必须出现，也就是要求是子集关系（`<ExpectedObject>` 是 `<SourceObject>` 的子集）。`<ExpectedObject>` 中某个变量如果还未绑定，则按照对应的 `<SourceObject>` 的值进行绑定；如果变量已经绑定，则两边的值必须是匹配的。

支撑我们整个 DSL 的基石，正是匹配文法。这里你可以把所有支持的命令都看成是 bool 表达式，如果返回 true 则成功，返回 false 则失败。我们看下一开始你看到的例子的片段：

```bash
ret 298
json '[{"code": $(code1)}, {"code": 612}]'
```

它表达的含义是，要求返回包的 StatusCode = 298，然后返回的 Response Body 必须能够匹配 `'[{"code": $(code1)}, {"code": 612}]'`，Content-Type 则必须为 `application/json`。它等价于：

```bash
ret		#不带参数的 ret 仅仅发起请求，并将返回包存储在 $(resp) 变量中，不做任何匹配
match 298 $(resp.code)
match '["application/json"]' $(resp.header.Content-Type)
match '[{"code": $(code1)}, {"code": 612}]' $(resp.body)
```

# 子命令

如同 linux shell 一样，我们可以在一条命令中，嵌入另一个命令，并把该命令的执行结果作为本命令输入的一部分。这种嵌入其他命令之中的命令，我们称为子命令。样例如下：

```bash
host rs.qiniu.com `env QiniuRSHost`

match $(ekey1) |base64 Bucket:Key|
match $(ekey2) |base64 Bucket2:Key2|
```

和 linux shell 相比，我们多了一个子命令语法：`|...|`。这没有别的意图，纯粹是为了 Go 语言的友好性（linux 风格的子命令在 Go 里面表达需要特别费劲）。

我们样例中的两个子命令 `env` 和 `base64` 都是返回 string 类型。但作为我们 DSL 的一部分，子命令同样可以返回我们类型系统中的任意类型。所以，原则上我们的子命令如同变量一样，有着上下文相关的 marshal 需求，比如：

```bash
post http://rs.qiniu.com/delete/`base64 Bucket:Key`
match $(foo) {"ekey":`base64 Bucket:Key`}
```

为了达到这样的效果，我们可以想象一种子命令的实现手法：

```bash
match $(__auto_var_1) `base64 Bucket:Key`
post http://rs.qiniu.com/delete/$(__auto_var_1)

match $(__auto_var_2) `base64 Bucket:Key`
match $(foo) {"ekey":$(__auto_var_2)}
```

也就是为每个子命令背地里生成一个自动变量，这样就可以让上下文相关的 marshal 能力，统一到由 “智能变量” 来支持。

# HTTP API 测试

请求包：

```bash
req <Method> <Url>      #可以简写为 post <Url> 或 get <Url> 或 delete <Url>
auth <AuthInterface>
header <Key1> <Value1>
header <Key2> <Value2>
body <BodyType> <Body>  #可以简写为 form <FormBody> 或 json <JsonBody>
```

返回包测试：

```bash
ret <Code>              #参数可不指定。不带参数的 ret 仅仅发起请求，并将返回包存储在 $(resp) 变量中
header <Key1> <Value1>
header <Key2> <Value2>
body <BodyType> <Body>  #可以简写为 json <JsonBody>
```

# 多案例支持

一般测试案例框架都有选择性执行某个案例、多个案例共享 setUp、tearDown 这样的启动和终止代码。我们 DSL 也支持，如下：

```bash
#代码片段1
...

case testCase1
#代码片段2
...

case testCase2
#代码片段3
...

tearDown
#代码片段4
...
```

这段代码里面，“代码片段1” 将被认为是 setUp 代码，“代码片段4” 是 tearDown 代码，所有 testCase 开始前都会执行一遍“代码片段1”，退出前执行一遍“代码片段4”。每个 case 不用写 end 语句，遇到下一个 case 或者遇到 tearDown 就代表该 case 结束。


# 运算能力

目前，这套 DSL 的运算能力是比较有限的。基本上只能做字符串拼接（concat）。如下：

```bash
match $(c) '"Hello $(a), $(b)!"'
```

如果我们希望做复杂运算，我设想未来有可能通过支持 `calc` 这样的子命令。例如：

```bash
match $(g) `calc max($(a), $(b), $(c)) + sin($(d)) + $(e)`
```

实现一个 `calc` 并不复杂，在 C++ 中用 [TPL](https://github.com/xushiwei/tpl) 只是几十分钟的事情（但在 Go 语言里面怎么做还没有特别去研究）。

考虑尽可能利用现有资源的话，我们可以考虑内嵌 lua 来实现 calc 支持。比如：

```bach
match $(g) `calc math.max($(a), $(b), $(c)) + math.sin($(d)) + $(e)`
```

参考：

* https://github.com/aarzilli/golua
* https://github.com/stevedonovan/luar (基于 golua 的进一步包装)

有了 `calc` 事情就更有意思了，我们还可以直接用 `calc` 命令做断言，比如：

```bash
calc $(a) < $(b)
```

在 `calc` 外面不套任何指令，由于 `calc` 返回 `false` 或 `true`，而基于前面返回 `true` 表示成功，返回 `false` 表示失败的原则，这个指令直接就是断言。当然为了友好，我们可以搞个别的名字：

```bash
assert $(a) < $(b)
```

# 流程控制

等等，难道我们真要做一个图灵完备的语言？上面的运算能力的讨论已经有点脱离需求了（先实际使用中检验吧），我们就此打住吧。
