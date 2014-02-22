(function() {
    // ---------------------------------------------------------- utils --- //
    var escapePath = function(path) {
        path = decodeURIComponent(path);
        var parts = path.split("/");
        var escapedPath = "";
        for (i in parts) {
            escapedPath += "/" + encodeURIComponent(parts[i]);
        }
        return escapedPath.substring(1);
    }

    var App = Backbone.Model.extend({
        defaults: function() {
            return {
                selected: null,
                _prevSelected: null,
            }
        },
        initialize: function() {
            this.bind('change:selected', this.selectedChanged, this);
            this.get('artists').bind('add', this.collectionItemAdded, this);
            this.get('albums').bind('add', this.collectionItemAdded, this);
        },
        collectionItemAdded: function(i) {
            var selected = this.attributes.selected;
            if (!selected)
                return;

            if (i.get('type') === selected.collection &&
                i.get('name') === selected.item) {
                i.set('selected', true);
            }
        },
        selectedChanged: function(e) {
            console.log('sel changed');
            console.log(e);
            var prev = this.attributes._prevSelected;
            var curr = this.attributes.selected;

            // unselect the previous selection
            if (prev) {
                var collection = this.get(prev.collection + 's'); // pluralize
                var items = collection.filter(function(item) {
                    return item.get('name') === prev.name;
                });
                if (items.length)
                    items[0].set('selected', false);
            }

            this.set('_prevSelected', curr);
        }
    });

    // --------------------------------------------------------- models --- //
    var Category = Backbone.Model.extend({
        // the API returns a string per album or artist
        parse: function(name) {
            return {
                name: decodeURIComponent(name)
            }
        },
        clear: function() {
            this.destroy();
        },
    });

    var Artist = Category.extend({
        defaults: function() {
            return {
                type: 'artist',
                selected: false
            }
        },
    });

    var Album = Category.extend({
        defaults: function() {
            return {
                type: 'album',
                selected: false
            }
        },
    });

    var Track = Backbone.Model.extend({
        defaults: function() {
            return {
                selected: false
            }
        },
        parse: function(track) {
            var path = '';
            return {
                title: decodeURIComponent(track.title),
                artist: decodeURIComponent(track.artist),
                album: decodeURIComponent(track.album),
                track: decodeURIComponent(track.track),
                path: escapePath(track.path)
            }
        },
        clear: function() {
            this.destroy();
        },
    });

    // ---------------------------------------------------- collections --- //
    var ArtistList = Backbone.Collection.extend({
        url: 'api/artist',
        model: Artist,
        comparator: function(a) {
            return a.get('name').toLowerCase();
        },
    });

    var AlbumList = Backbone.Collection.extend({
        url: 'api/album',
        model: Album,
        comparator: function(a) {
            return a.get('name').toLowerCase();
        },
    });

    var TrackList = Backbone.Collection.extend({
        model: Track,
        initialize: function(args) {
            this.collections = args.collections;
        },
        load: function(m) {
            /*
            var collections = this.collections;
            // clear any existing highlighted artist or album
            for (k in collections) {
                if (collections[k].curr) {
                    collections[k].curr.set('selected', false);
                    collections[k].curr = null;
                }
            }
            collections[m.get('type')].curr = m;*/
            m.set('selected', true);
            this.url = 'api/' + m.get('type') + '/' + encodeURIComponent(m.get('name'));
            this.fetch({reset: true});
        }
    });

    // ---------------------------------------------------------- views --- //
    var CategoryView = Backbone.View.extend({
        tagName: 'div',
        template: _.template($('#list-template').html()),
        events: {
            'click a': 'click'
        },
        initialize: function(args) {
            this.app = args.app;
            this.model.bind('change', this.render, this);
        },
        curr: null,
        click: function(e) {
            //this.model.set('selected', true);
        },
        render: function() {
            this.$el.html(this.template(this.model.toJSON()));
            return this;
        }
    });

    var TrackView = Backbone.View.extend({
        tagName: 'div',
        template: _.template($('#track-template').html()),
        events: {
            'click a.track': 'play'
        },
        initialize: function(args) {
            this.nowPlaying = args.nowPlaying;
            this.model.bind('change', this.render, this);
        },
        render: function() {
            this.$el.html(this.template(this.model.toJSON()));
            return this;
        },
        play: function(e) {
            e.preventDefault();
            this.nowPlaying.start(this.model);
        }
    });

    var NowPlayingView = Backbone.View.extend({
        el: $('#now-playing'),

        events: {
            'canplay': 'canPlay',
            'ended': 'ended'
        },

        shouldPlay: true,
        prev: null,
        curr: null,

        initialize: function(args) {
            this.tracks = args.tracks;
            this.$el.hide();
        },
        start: function(track, startPaused) {
            if (this.curr) {
                this.prev = this.curr;
            }
            this.curr = track;
            this.$el.show();
            this.el.src = '/music/' + track.get('path');
            this.shouldPlay = !startPaused;
        },
        canPlay: function() {
            if (this.shouldPlay)
                this.el.play();
            if (this.prev) {
                this.prev.set('selected', false);
                this.prev = null;
            }
            this.curr.set('selected', true);
        },
        togglePause: function() {
            if (!this.el.src)
                return;
            if (this.shouldPlay)
                this.el.pause();
            else
                this.el.play();
            this.shouldPlay ^= true;
        },
        playPrev: function() {
            if (!this.curr)
                return;
            var found = false;
            var i;
            for (i in Tracks.models) {
                if (this.curr === this.tracks.models[i]) {
                    found = true;
                    if (i > 0)
                        i--;
                    break;
                }
            }
            if (!found && this.tracks.models.length) {
                // didn't find it.  must have switched categories.
                // start playing the first track.
                this.start(this.tracks.models[0], !this.shouldPlay);
                return true;
            } else {
                // found one. play it.
                if (this.el.currentTime < 5)
                    this.start(Tracks.models[i], !this.shouldPlay);
                else
                    this.el.currentTime = 0;
                return true;
            }
            return false;
        },
        playNext: function() {
            var found = false;
            var i;
            for (i in this.tracks.models) {
                if (this.curr === this.tracks.models[i]) {
                    found = true;
                    i++;
                    break;
                }
            }
            if (!found && this.tracks.models.length) {
                // didn't find it.  must have switched categories.
                // start playing the first track.
                this.start(this.tracks.models[0], !this.shouldPlay);
                return true;
            } else if (i < this.tracks.models.length) {
                // found one. play it.
                this.start(this.tracks.models[i], !this.shouldPlay);
                return true;
            }
            return false;
        },
        ended: function() {
            // current song is over.  see if there is another one to
            // play.
            if (!this.playNext()) {
                // end of the track list. clean up.
                this.shouldPlay = true;
                this.curr.set('selected', false);
                this.prev = null;
                this.curr = null;
                this.$el.hide();
            }
        }
    });

    var AppView = Backbone.View.extend({
        el: $('#content'),
        initialize: function(args) {
            this.nowPlaying = args.nowPlaying;

            this.model.bind('change', this.selectionChanged, this);

            this.model.get('artists').bind('add', this.addOne, this);
            this.model.get('artists').bind('reset', this.addAllArtists, this);

            this.model.get('albums').bind('add', this.addOne, this);
            this.model.get('albums').bind('reset', this.addAllAlbums, this);

            this.model.get('tracks').bind('add', this.addTrack, this);
            this.model.get('tracks').bind('reset', this.addAllTracks, this);

            this.model.get('artists').fetch();
            this.model.get('albums').fetch();
        },
        addOne: function(a) {
            var view = new CategoryView({model: a, app: this.model});
            var type = a.get('type');
            $('#' + type).append(view.render().el);
        },
        addAllArtists: function() {
            this.artists.each(this.addOne, this);
        },
        addAllAlbums: function() {
            this.albums.each(this.addOne, this);
        },
        addTrack: function(t) {
            var view = new TrackView({model: t, nowPlaying: this.nowPlaying});
            $('#tracks').append(view.render().el);
        },
        addAllTracks: function() {
            $('#tracks').empty();
            this.tracks.each(this.addTrack, this);
        },
        selectionChanged: function(e) {
            console.log('selection changed');
            console.log(e)
        },
    });

    var AppRouter = Backbone.Router.extend({
        initialize: function(args) {
            this.app = args.app;
        },
        routes: {
            '': 'overview',
            ':collection/:item': 'details',
        },
        'overview': function() {
            this.app.set('selected', null);
        },
        'details': function(collection, item) {
            this.app.set('selected', {
                collection: collection,
                item: item,
            });
        },
    });

    // constants
    var codeSpace = ' '.charCodeAt();
    var codeArrowRight = 39;
    var codeArrowLeft = 37;

    // ------------------------------------------------- initialization --- //
    $(function() {
        var artists = new ArtistList();
        var albums = new AlbumList();
        var collections = {
            'artist': artists,
            'album': albums
        };
        var tracks = new TrackList({collections: collections});
        var nowPlaying = new NowPlayingView({tracks: tracks});
        var app = new App({
            artists: artists,
            albums: albums,
            tracks: tracks,
        });
        var appView = new AppView({
            model: app,
            nowPlaying: nowPlaying,
        });
        var router = new AppRouter({app: app});
        Backbone.history.start();

        $(document).delegate('a', 'click', function(e) {
            var a = $(this).attr('href');
            if (a.length > 0 && a[0] === '/' && a.slice(2) !== '//') {
                e.preventDefault();
                Backbone.Router.navigate(a, {trigger: true});
            }
        });

        // global key handlers for playback control
        $(document).bind('keyup', function(e) {
            switch(e.which) {
            case codeArrowRight:
                e.preventDefault();
                nowPlaying.playNext();
                break;
            case codeArrowLeft:
                e.preventDefault();
                nowPlaying.playPrev();
                break;
            case codeSpace:
                e.preventDefault();
                if (nowPlaying.curr)
                    nowPlaying.togglePause();
                else
                    nowPlaying.playNext();
                break;
            }
        });

        // warn if using firefox, because it doesn't have aac or mp3 codecs
        if (navigator.userAgent.indexOf('Chrome') < 0) {
            $('#footer').append('<span style="color: red; font-weight: bold;">(Firefox can\'t play aac or mp3 files.  Use Chrome.)</span>');
        }
    });
})();
