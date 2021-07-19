class SearchData extends Atom {

}


class SearchResults extends CollectionStoreOf(SearchData) {
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
                        console.log("hello?", result);
                        this.setStore(result.map(element => new SearchData(element)));
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
            ${title}
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
    init(router) {
        this.router = router;
        this.searchInput = "";
        this.searchData = new SearchResults();
        this.searchResultsList = new SearchResultsList(this.searchData);
        this.handleInput = this.handleInput.bind(this);
    }

    handleInput(evt) {
        this.searchInput = evt.target.value;
        //get search results
        //note don't need to call render since this list will re-render once the collection store updates
        this.searchData.fetch(this.searchInput)
                        .then()
                        .catch(ex => {
                            //if an error occured, page won't render so need to call render to update with error message
                            this.render();
                        }) 
    }

    create() {
        return html`<div class = "engine">
            <h1>Search</h1>
            <input oninput=${this.handleInput} value=${this.searchInput} placeholder="Search across your digital footprint"/>
            ${this.searchResultsList.node}
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
        this.handleLink = (evt) => this.handleLink("link", evt);
        this.handleContent = (evt) => this.handleContent("content", evt);
        this.handleTags = (evt) => this.handleTags("tags", evt);
    }

    handleInput(el, evt) {
        this.data.update({[el]: evt.target.value})
    }

    create({title, link, content, tags}) {
        return html`<div class="footprint">
            <div class="content">
                <input oninput=${this.handleTitle} value=${title}/> 
                <input oninput=${this.handleLink} value=${link}/> 
                <input oninput=${this.handleTags} value=${tags}/> 
                <div class = "datacontent">
                    <textarea oninput=${this.handleContent} class="littlePadding" placeholder="Paste some content" value=${content}></textarea>
                    <pre class="p-heights littlePadding ${content.endsWith("\n") ? 'endline' : ''}">${content}</pre>
                </div>
            </div>
        </div>`
    }
}


class App extends Component {
    init() {
        this.engine = new SearchEngine();
        this.router = new Router();
        this.router.on({
            route: "/",
            handler: (route) =>{
                this.route = route;
                //do some stuff to get the route to render
            }
        })
    }

    create() {
        return html`<main>
            <div class = "content">
                ${this.engine.node}
            </div>
            <footer>Built with <a href="https://github.com/amirgamil/poseidon">Poseidon</a> by <a href="https://amirbolous.com/">Amir</a></footer>
        </main>` 
    }
}

const app = new App();
document.body.appendChild(app.node);