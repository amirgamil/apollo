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

//takes a string of content and returns
//a text with HTML tags injected for key query words
const highlightContent = (text, query) => {
    const regex = new RegExp(query.join(' '));
    return text.replace(regex, `<span class="highlighted">${query[0]}</span>`);
}


class SearchResults extends CollectionStoreOf(Data) {
    fetch(query) {
        return fetch("/search?q=" + encodeURIComponent(query), 
                {
                    method: "POST",
                    mode: "no-cors",
                    // headers: {
                    //     "Accept-Encoding": "gzip, deflate"
                    // },
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
                        this.setStore(result.data.map((element, id) => {
                            element["selected"] = id === 0 ? true : false;
                            element["content"] = highlightContent(element["content"], this.query); 
                            return new Data(element);
                        }));
                    } else {
                        this.setStore([]);
                    }
               })
               .catch(ex => {
                   console.log("Exception occurred trying to fetch the result of a request: ", ex);
               })
    }
}

class Result extends Component {
    init(data, removeCallBack) {
        this.data = data;
        this.removeCallBack = removeCallBack;
        this.displayDetails = false;
        this.loadPreview = this.loadPreview.bind(this);
        this.closeModal = this.closeModal.bind(this);
        this.bind(data);
    }

    loadPreview() {
        //fetch the full text
        fetch("/getRecordDetail?q=" + this.data.get("title"), {
            method: "POST",
            mode: "no-cors",
            body: JSON.stringify()
        }).then(data => data.json())
          .then(res => {
              this.displayDetails = true;
              //add highlighting
              this.data.update({"fullContent": res});
          }).catch(ex => {
              console.log("Error fetching details of item: ", ex);
          })
    }

    closeModal(evt) {
        //stop bubbling up DOM which would cancel this action by loading preview
        evt.stopPropagation();
        this.displayDetails = false;
        this.render();
    }


    create({title, link, content, selected, fullContent}) {
        const contentToDisplay = content + "..."
        return html`<div class="result colWrapper ${selected ? 'hoverShow' : ''}" onclick=${this.loadPreview}>
            <a onclick=${(evt) => evt.stopPropagation()} href=${link}>${title}</a>
            <p innerHTML = ${contentToDisplay}></p>
            ${this.displayDetails ? html`<div class = "modal"> 
                    <div class="modalContent">
                        <div class="windowBar">
                            <p class="modalNavTitle">details.txt</p> 
                            <div class="navPattern"></div>
                            <button class="closeModal" onclick=${this.closeModal}>x</button>
                        </div>
                        <div class="modalBody"> 
                            <div class="rowWrapper"> 
                                <h2>${title}</h2>
                            </div>
                            <p><a href=${link}>Source</a></p>
                            <p innerHTML = ${fullContent}></p>
                        </div>
                    </div>
                </div>` : null}
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
        this.modalT
        //used to change selected results based on arrow keys
        this.selected = 0;
        this.time = ""
        //add a little bit of delay before we search because too many network requests
        //will slow down retrieval of search results, especially as user is typing to their deired query
        //each time the user lifts up their finger from the keyboard, debounce will fire which will
        //check if 500ms has elapsed, if it has, will query and load the search results,
        //otherwise if it's called again, rinse and repeat
        this.loadSearchResults = debounce(this.loadSearchResults.bind(this), 500);
        this.setSearchInput = this.setSearchInput.bind(this);
        this.handleKeydown = this.handleKeydown.bind(this);
        this.toggleSelected = this.toggleSelected.bind(this);
        //if we have a query on initialization, navigate to it directly
        if (this.query) {
            this.setSearchInput(this.query);
            this.loadSearchResults(this.query);
        }
    }

    //TODO: add pagination into API to return e.g. 20 results and load more for speed
    loadSearchResults(evt) {
        if (evt.key === "ArrowDown" || evt.key === "ArrowUp" || evt.key === "Enter" || evt.key === "Escape") {
            return ;
        } 
        this.searchData.fetch(this.searchInput)
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
        // this.loadSearchResults(this.searchInput);
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

    toggleSelected(state) {
        const listSize = this.searchResultsList.size;
        switch (state) {
            case "ArrowDown":
                this.selected += 1
                if (this.selected < listSize) {
                    window.scrollBy(0, 100);
                    this.searchResultsList.nodes[this.selected - 1].data.update({"selected": false});
                    this.searchResultsList.nodes[this.selected].data.update({"selected": true});
                } else {
                    window.scrollTo(0, 0);
                    this.selected = 0;
                    this.searchResultsList.nodes[this.selected].data.update({"selected": true});
                    this.searchResultsList.nodes[listSize - 1].data.update({"selected": false});
                }
                break;
            case "ArrowUp":
                this.selected -= 1;
                if (this.selected >= 0) {
                    window.scrollBy(0, -100);
                    this.searchResultsList.nodes[this.selected + 1].data.update({"selected": false});
                    this.searchResultsList.nodes[this.selected].data.update({"selected": true});
                } else {
                    window.scrollBy(0, document.body.scrollHeight);
                    this.selected = listSize - 1;
                    this.searchResultsList.nodes[0].data.update({"selected": false});
                    this.searchResultsList.nodes[this.selected].data.update({"selected": true});
                }
            
        }
        this.searchResultsList.nodes[this.selected].render();
    }

    handleKeydown(evt) {
        //deal with cmd a + backspace should empty all search results
        if (evt.key === "ArrowDown" || evt.key === "ArrowUp") {
            //change the selected attribute
            evt.preventDefault();
            this.toggleSelected(evt.key);
        } else if (evt.key === "Enter") {
            evt.preventDefault();
            this.searchResultsList.nodes[this.selected].loadPreview();
        } else if (evt.key === "Escape") {
            evt.preventDefault();
            this.searchResultsList.nodes[this.selected].displayDetails = false;
            this.searchResultsList.nodes[this.selected].render();
        }
    }

    create() {
        const time = this.searchData.time ? this.searchData.time.toFixed(2) : 0
        return html`<div class = "engine">
            <h1 class="engineTitle"><span class="blue">A</span><span class="red">p</span><span class="yellow">o</span><span class="blue">l</span><span class="green">l</span><span class="yellow">o</span></h1>
            <input onkeydown=${this.handleKeydown} oninput=${this.handleInput} onkeyup=${this.loadSearchResults} value=${this.searchInput} placeholder="Search my digital footprint"/>
            <p class="time">${this.searchInput ? "About " + this.searchData.size + " results (" + time + "ms)" : html`<p>To navigate with your keyboard: <strong>Arrow keys</strong> move up and down results, <strong>Enter</strong> opens the result in detail, <strong>Escape</strong>
            closes the detail view</p>`}</p>
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
        this.showModal = false;
        this.modalText = "";
        this.handleContent = (evt) => this.handleInput("content", evt);
        this.handleTags = (evt) => this.handleInput("tags", evt);
        this.addData = this.addData.bind(this);
        this.scrapeData = this.scrapeData.bind(this);
        this.closeModal = this.closeModal.bind(this);
        this.password = "";
        this.isAuthenticated = window.localStorage.getItem("authenticated") === "true";
        this.authenticatePassword = this.authenticatePassword.bind(this);
        this.showAuthError = this.showAuthError.bind(this);
        this.updatePassword = this.updatePassword.bind(this);
        this.bind(this.data);
    }

    authenticatePassword() {
        fetch("/authenticate", {
            method: "POST",
            mode: "no-cors",
            headers: {
                "Content-Type": "application/json"
            }, 
            body: JSON.stringify({
                "password": this.password
            })
        }).then(response => {
            if (response.ok) {
                window.localStorage.setItem("authenticated", "true");
                this.modalText = "Hooray!"
                this.showModal = true;
                this.render();
            } else {
                window.localStorage.getItem("authenticated", "false");
                return Promise.reject(response);
            }
        }).catch(e => {
            this.showAuthError();
            return;
        })
    }

    showAuthError() {
        this.modalText = "You're not Amir :("
        this.showModal = true;
        this.render();
    }

    closeModal() {
        this.showModal = false;
        this.render();
    }

    updatePassword(evt) {
        this.password = evt.target.value;
        this.render();
    }


    scrapeData(evt) {
        if (!this.isAuthenticated) {
            this.showAuthError();
            return;
        }
        this.showModal = true;
        this.modalText = "Hold on, doing some magic..."
        this.render();
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
            this.showModal = false;
            this.data.update({title: data["title"], content: data["content"]});
            // window.scrollBy(0, document.body.scrollHeight);
        }).catch(ex => {
            console.log("Exception trying to fetch the article: ", ex)
            this.modalText = "Error scraping, sorry!";
            this.render();
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
        if (!this.isAuthenticated) {
            this.showAuthError();
            return;
        }
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
                this.showModal = true;
                this.modalText = "Success!"
                this.render();
            } else {
                Promise.reject(response)
            }
        }).catch(ex => {
            console.log("Error adding to the db: ", ex);
            this.showModal = true;
            this.modalText = "Error scraping, sorry!";
            this.render();
        })
    }

    handleInput(el, evt) {
        this.data.update({[el]: evt.target.value})
    }

    create({title, link, content, tags}) {
        console.log(content);
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
            <h3>Are you Amir? Please prove yourself</h3>
            <input oninput=${this.updatePassword} value=${this.password} type="password" placeholder="I am not :(" />
            <button class="action" onclick=${this.authenticatePassword}>Prove</button>
            ${this.showModal ? html`<div class = "modal"> 
                    <div class="modalContent">
                        <div class="windowBar">
                            <p class="modalNavTitle">popup</p> 
                            <div class="navPattern"></div>
                            <button class="closeModal" onclick=${this.closeModal}>x</button>
                        </div>
                        <div class="modalBody"> 
                            <p>${this.modalText}</p>
                        </div>
                    </div>
                </div>` : null}
        </div>`
    }
}

const about = html`<div class="colWrapper">
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
        this.router = new Router(3);
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
            <footer>Built with <a href="https://github.com/amirgamil/poseidon">Poseidon</a> by <a href="https://amirbolous.com/">Amir</a> and <a href="https://github.com/amirgamil/poseidon">open source</a> on GitHub</footer>
        </main>` 
    }
}

const app = new App();
document.body.appendChild(app.node);