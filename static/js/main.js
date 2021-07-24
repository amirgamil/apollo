// only fire fn once it hasn't been called in delay ms
const debounce = (fn, delay) => {
    let to = null;
    return (...args) => {
        const bfn = () => fn(...args);
        clearTimeout(to);
        to = setTimeout(bfn, delay);
    }
}

class Data extends Atom {

}


class SearchResults extends CollectionStoreOf(Data) {
    fetch(query) {
        return fetch("/search?q=" + encodeURIComponent(query), 
                {
                    method: "POST",
                    mode: "no-cors",
                    body: JSON.stringify()
                })
               .then(response => {
                   if (response.ok) {
                       return response.json();
                   } else {
                       Promise.reject(response);
                   }
               }).then(result => {
                    if (result) {
                        //time comes back in nanoseconds
                        this.time = result.time * 0.000001;
                        this.query = result.query;
                        this.setStore(result.data.map(element => new Data(element)));
                    } else {
                        this.setStore([]);
                    }
               }).catch(ex => {
                   console.log("Exception occurred trying to fetch the result of a request: ", ex);
               })
    }
}

class Result extends Component {
    init(data, removeCallBack) {
        this.data = data;
        this.removeCallBack = this.removeCallBack;
    }

    create({title, link, content}) {
        return html`<div>
            <a href=${link}>${title}</a>
            <p>${content.slice(0, 100) + "..."}</p>
        </div>`
    }
}

class SearchResultsList extends ListOf(Result) {
    create() {
        return html`<div class="colWrapper">
            ${this.nodes}
        </div>`;
    }
}


class SearchEngine extends Component {
    init(router, query) {
        this.router = router;
        this.query = query;
        this.searchInput = "";
        this.searchData = new SearchResults();
        this.searchResultsList = new SearchResultsList(this.searchData);
        this.handleInput = this.handleInput.bind(this);
        this.loading = false;
        this.time = ""
        //add a little bit of delay before we search because too many network requests
        //will slow down retrieval of search results, especially as user is typing to their deired query
        this.loadSearchResults = debounce(this.loadSearchResults.bind(this), 100);
        this.setSearchInput = this.setSearchInput.bind(this);
        //if we have a query on initialization, navigate to it directly
        if (this.query) {
            this.setSearchInput(this.query);
            this.loadSearchResults(this.query);
        }
    }

    loadSearchResults(value) {
        this.searchData.fetch(value)
                        .then(() => {
                            this.loading = false;
                            this.render();
                        })
                        .catch(ex => {
                            //if an error occured, page won't render so need to call render to update with error message
                            this.render();
                        }) 
    }

    setSearchInput(value) {
        this.searchInput = value;
    }

    handleInput(evt) {
        this.setSearchInput(evt.target.value);
        this.router.navigate("/search?q=" + encodeURIComponent(evt.target.value));
        this.loading = true;
        this.render();
        //get search results
        this.loadSearchResults(this.searchInput);
    }

    styles() {
        return css`
            .engineTitle {
                align-self: center;
            }
            .blue {
                color: #2A63BF;
            }

            .red {
                color: #E34133;
            }
            .yellow {
                color: #F3B828;
            }
            .green {
                color: #32A556;
            }
        `
    }

    create() {
        const time = this.searchData.time ? this.searchData.time.toFixed(2) : 0
        return html`<div class = "engine">
            <h1 class="engineTitle"><span class="blue">A</span><span class="red">p</span><span class="yellow">o</span><span class="blue">l</span><span class="green">l</span><span class="yellow">o</span></h1>
            <input oninput=${this.handleInput} value=${this.searchInput} placeholder="Search my digital footprint"/>
            <p class="time">${this.searchInput ? "About " + this.searchData.size + " results (" + time + "ms)" : null}</p>
            ${this.loading ? html`<p>loading...</p>` : this.searchResultsList.node} 
        </div>`
    }
}

//where we add data for now, probably going to change
class DigitalFootPrint extends Component {
    init() {
        //initalize stuff here
        this.data = new Data({title: "", link: "", content: "", tags: ""})
        this.handleInput = this.handleInput.bind(this);
        this.handleTitle = (evt) => this.handleInput("title", evt);
        this.handleLink = (evt) => this.handleInput("link", evt);
        this.handleContent = (evt) => this.handleInput("content", evt);
        this.handleTags = (evt) => this.handleInput("tags", evt);
        this.addData = this.addData.bind(this);
        this.scrapeData = this.scrapeData.bind(this);
        this.bind(this.data);
    }

    scrapeData(evt) {
        fetch("/scrape?q=" + this.data.get("link"), {
            method: "POST",
            mode: "no-cors",
            headers: {
                "Content-Type" : "application/json"
            },
        }).then(response => {
            if (response.ok) {
                return response.json()
            } else {
                Promise.reject(response)
            }
        }).then(data => {
            this.data.update({title: data["title"], content: data["content"]});
        }).catch(ex => {
            console.log("Exception trying to fetch the article: ", ex)
        })
    }

    getTagArrayFromString(tagString) {
        //remove whitespace
        tagString = tagString.replace(/\s/g, "");
        let tags = tagString.split('#');
        tags = tags.length > 1 ? tags.slice(1) : [];
        return tags;
    }

    addData() {
        //create array from text tags
        let tags = this.getTagArrayFromString(this.data.get("tags"));
        fetch("/addData", {
            method: "POST",
            mode: "no-cors",
            headers: {
                "Content-Type":"application/json"
            },
            body: JSON.stringify({
                title: this.data.get("title"),
                link: this.data.get("link"),
                content: this.data.get("content"),
                tags: tags
            })
        }).then(response => {
            if (response.ok) {
                //TODO: change to actually display
                console.log("success!")
            } else {
                Promise.reject(response)
            }
        }).catch(ex => {
            console.log("Error adding to the db: ", ex);
        })
    }

    handleInput(el, evt) {
        this.data.update({[el]: evt.target.value})
    }

    create({title, link, content, tags}) {
        return html`<div class="colWrapper">
            <h1>Add some data</h1>
            <input oninput=${this.handleTitle} value=${title} placeholder="Title"/> 
            <input oninput=${this.handleLink} value=${link} placeholder="Link"/> 
            <input oninput=${this.handleTags} value=${tags} placeholder="#put #tags #here"/> 
            <div class = "datacontent">
                <textarea oninput=${this.handleContent} class="littlePadding" placeholder="Paste some content or scrape it" value=${content}></textarea>
                <pre class="p-heights littlePadding ${content.endsWith("\n") ? 'endline' : ''}">${content}</pre>
            </div>
            <div class="rowWrapper">
                <button class="action" onclick=${this.scrapeData}>Scrape</button> 
                <button class="action" onclick=${this.addData}>Add</button>
            </div>
        </div>`
    }
}

const about = html`<div>
    <h1>About</h1>
    <p>Apollo is an attempt at making something that has felt impersoal for the longest time, personal again. 
        
    </p>

    <p>
        The computer revolution produced
        <strong>personal computers</strong> yet <strong>impersonal search engines.</strong> So what's Apollo? It's a Unix-style search engine
        for your digital footprint. The design authentically steals from the past. This is intentional. When I use Apollo, I want to feel like I'm 
        <strong>travelling through the past.</strong>
    </p>

    <p>How do I define <strong>digital footprint?</strong> There are many possible definitions here, I define it as <strong>anything
        digital I come across that I want to remember in the future.</strong>
        
    </p>
    <p>
        It's like an indexable database or search engine for anything interesting I come across the web. There are also some personal data 
        sources I pull from like <a href="https://github.com/amirgamil/athena">Athena</a> for my thoughts or 
        <a href="https://zeus.amirbolous.com/">Zeus</a> for curated resources or <a href="https://www.amazon.com/b/?node=11627044011">Kindle Highlights</a>. 
        This is in addition to any interesting thing I come across the web, which I can add <a href = "">directly</a> via the web crawler. 
    </p>

    <p>The web crawler can scrape any article or blog post and reliably get the text - so you can index the <strong>entire post</strong> without even 
       having to copy it! Once again, this is intentional. I read a lot of stuff on the Internet but don't take notes (because I'm lazy). Now I can 
       index <strong>anything interesting I come across</strong> and don't have to feel guilty about not having made notes. So just to be clear, 
       I'm not indexing just the name of an article - I'm indexing <strong>the entire contents!</strong> If that's not cool, I don't know what is.
    </p>

    <p>I no longer have to rely on my memory to index anything interesting I come across. And now you don't have to either</p>

    <p>P.S I put a lot of ❤️ in this project, I hope you like it :) </p>

</div>`

class App extends Component {
    init() {
        this.router = new Router();
        this.footprint = new DigitalFootPrint();
        this.router.on({
            route: "/search",
            handler: (route, params) => {
                this.engine = new SearchEngine(this.router, params["q"]);
                this.route = route;
                this.render();
            }
        });

        this.router.on({
            route: ["/about", "/add"],
            handler: (route) => {
                this.route = route;
                this.render();
            }
        })

        this.router.on({
            route: "/",
            handler: (route) => {
                this.engine = new SearchEngine(this.router);
                this.route = route;
                this.render();
            }
        });
    }

    create() {
        const hour = new Date().getHours();
		if (hour > 19 || hour < 7) {
			document.body.classList.add('dark');
			document.documentElement.style.color = '#222';
		} else {
			document.body.classList.remove('dark');
			document.documentElement.style.color = '#fafafa';
		}
        return html`<main>
            <nav>
                <div class="topNav"> 
                    <h5 class="titleNav"><strong class="cover">Apollo: A personal 🔎 engine</strong></h5>
                    <h5 class="welcomeNav">Amir's Digital 👣</h5>
                    <div class="navSubar">
                        <button title="Home" onclick=${() => this.router.navigate("/")}><img src="static/img/home.png" /></button>
                        <button title="Add a record" onclick=${() => this.router.navigate("/add")} ><img src="static/img/add.png" /></button>
                        <button title="About" onclick=${() => this.router.navigate("/about")} ><img src="static/img/about.png" /></button>
                        <input class="navInput" placeholder=${window.location.href} />
                    </div>
                </div>
            </nav>
            <div class = "content">
                ${() => {
                    switch (this.route) {
                        case "/add":
                            return this.footprint.node;
                        case "/about":
                            return about;
                        default:
                            return this.engine.node;
                    }
                }}
            </div>
            <footer>Built with <a href="https://github.com/amirgamil/poseidon">Poseidon</a> by <a href="https://amirbolous.com/">Amir</a></footer>
        </main>` 
    }
}

const app = new App();
document.body.appendChild(app.node);