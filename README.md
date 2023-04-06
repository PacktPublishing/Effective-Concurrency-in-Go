# Effective Concurrency in Go

<a href="https://www.amazon.com/dp/1804619078"><img src="" alt="Effective Concurrency in Go" height="256px" align="right"></a>

This is the code repository for [Effective Concurrency in Go](https://www.amazon.com/dp/1804619078), published by Packt.

**Develop, analyze, and troubleshoot high performance concurrent applications with ease**

## What is this book about?
The Go language has been gaining momentum due to its treatment of concurrency as a core language feature, making concurrent programming more accessible than ever. However, concurrency is still an inherently difficult skill to master, since it requires the development of the right mindset to decompose problems into concurrent components correctly. This book will guide you in deepening your understanding of concurrency and show you how to make the most of its advantages.

This book covers the following exciting features:
* Understand basic concurrency concepts and problems
* Learn about Go concurrency primitives and how they work
* Learn about the Go memory model and why it is important
* Understand how to use common concurrency patterns
* See how you can deal with errors in a concurrent program
* Discover useful techniques for troubleshooting

If you feel this book is for you, get your [copy](https://www.amazon.com/dp/1804619078) today!

<!--- <a href="https://www.packtpub.com/?utm_source=github&utm_medium=banner&utm_campaign=GitHubBanner"><img src="https://raw.githubusercontent.com/PacktPublishing/GitHub/master/GitHub.png" alt="https://www.packtpub.com/" border="5" /></a> --->

## Instructions and Navigations
All of the code is organized into folders. For example, Chapter02.

The code will look like the following:
```
1: chn := make(chan bool) // Create an unbuffered channel 
2: go func() { 
3: chn <- true // Send to channel 
4: }() 
5: go func() { 
6: var y bool 
7: y <-chn // Receive from channel 
8: fmt.Println(y) 
9: }() 
```

**Following is what you need for this book:**
If you are a developer with basic knowledge of Go and are looking to gain expertise in highly concurrent backend application development, then this book is for you. Intermediate Go developers who want to make their backend systems more robust and scalable will also find plenty of useful information. Prior exposure Go is a prerequisite.

With the following software and hardware list you can run all code files present in the book (Chapter 1-10).
### Software and Hardware List
| Chapter | Software required | OS required |
| -------- | ------------------------------------ | ----------------------------------- |
| 1-10 | Latest version of GO | Windows, Mac OS X, and Linux (Any) |


We also provide a PDF file that has color images of the screenshots/diagrams used in this book. [Click here to download it](https://packt.link/3rxJ9).

### Related products
*  Functional Programming in Go [[Packt]](https://www.packtpub.com/product/functional-programming-in-go/9781801811163?utm_source=github&utm_medium=repository&utm_campaign=9781801811163) [[Amazon]](https://www.amazon.com/dp/1801811164)

* Domain-Driven Design with Golang [[Packt]](https://www.packtpub.com/product/domain-driven-design-with-golang/9781804613450?utm_source=github&utm_medium=repository&utm_campaign=9781804613450) [[Amazon]](https://www.amazon.com/dp/1804613452)


## Get to Know the Author
**Burak Serdar**
is a software engineer with over 30 years of experience in designing and developing distributed enterprise applications that scale. He’s worked for several startups and large corporations like Thomson and Red Hat as an engineer and technical lead.

He’s one of the co-founders of Cloud Privacy Labs where he works on semantic interoperability and privacy technologies for centralized and decentralized systems. He tries to answer at least one question on StackOverflow every day. Burak holds BSc and MSc degrees in Electrical and Electronics Engineering, and an MSc degree in Computer Science.
