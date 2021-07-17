class Data extends Atom {

}


class SearchResults extends CollectionStoreOf(Data) {

}

class SearchResultsList extends List {

}


class SearchEngine extends Component {
    init() {
        this.searchData = new SearchResults();
        this.searchResultsList = new SearchResultsList(this.searchData);
    }

    create() {
        return html`<div class = "engine">
            ${this.searchResultsList.node}
        </div>`
    }
}

//where we add data for now, probably going to change
class DigitalFootPrint extends Component {
    init() {
        //initalize stuff here
        this.data = new Data({title: "", link: "", content: "", tags: ""})
    }

    create() {
        return html`<div class="footprint">
            <div class="content">
                <input /> 
                <input /> 
                <input />
                <div class = "datacontent">

                </div>
            </div>
        </div>`
    }
}


class App extends Component {
    init() {
    }

    create() {
        return html`<main>
            <div class = "content">

            </div>
            <footer>Built with <a href="https://github.com/amirgamil/poseidon">Poseidon</a> by <a href="https://amirbolous.com/">Amir</a></footer>
        </main>` 
    }
}

const app = new App();
document.body.appendChild(app.node);