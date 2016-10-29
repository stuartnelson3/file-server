import Html exposing (..)
import Html.App as Html
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as Json exposing (..)
import Task
import Navigation
import Debug
import UrlParser exposing (Parser, (</>), format, int, oneOf, s, string)
import String


main =
  Navigation.program urlParser
    { init = init
    , view = view
    , update = update
    , urlUpdate = urlUpdate
    , subscriptions = subscriptions
    }


-- UrlParsing


parse : Navigation.Location -> Route
parse {pathname} =
  let
      one = Debug.log "parse: path" pathname
      path =
        if String.startsWith "/" pathname then
          String.dropLeft 1 pathname
        else
          pathname
  in
     case UrlParser.parse identity routeParser path of
       Err err -> NotFound

       Ok route -> route


urlParser : Navigation.Parser Route
urlParser =
  Navigation.makeParser parse


moviesParser : Parser a a
moviesParser =
  UrlParser.oneOf
    [ (UrlParser.s "movies")
    , (UrlParser.s "")
    ]


movieParser : Parser (String -> a) a
movieParser =
  UrlParser.s "movie" </> UrlParser.string


routeParser : Parser (Route -> a) a
routeParser =
  UrlParser.oneOf
    [ format Movies moviesParser
    , format Movie movieParser
    ]


-- Model


type alias Model =
  { movies : List ApiResponse
  , movie : ApiResponse
  , route : Route
  }

init : Route -> (Model, Cmd Msg)
init route =
  urlUpdate route (Model [] (ApiResponse "" "" (MovieResponse "" "" "" "" "")) route)

type Route
  = Movies
  | Movie String
  | NotFound

-- Update


type Msg
  = Search
  | MoviesFetchSucceed (List ApiResponse)
  | MovieFetchSucceed ApiResponse
  | FetchFail Http.Error


update : Msg -> Model -> (Model, Cmd Msg)
update msg model =
  case msg of
    Search ->
      (model, searchApi)

    MoviesFetchSucceed movies ->
        ({ model | movies = movies }, Cmd.none)

    MovieFetchSucceed movie ->
        ({ model | movie = movie }, Cmd.none)

    FetchFail _ ->
      ({model | route = NotFound }, Cmd.none)


urlUpdate : Route -> Model -> (Model, Cmd Msg)
urlUpdate route model =
  let
      cmd =
        case route of
          Movie imdbID ->
            let
              one = Debug.log "urlUpdate: imdbID" imdbID
            in
              singleMovieSearch imdbID

          Movies ->
            searchApi

          _ ->
            Cmd.none
  in
    ({model | route = route }, cmd)


-- View


view : Model -> Html Msg
view model =
  case model.route of
    Movies ->
      moviesView model.movies

    Movie name ->
      let
        one = Debug.log "view: name" name
      in
        movieView model.movie

    _ ->
      notFoundView model

notFoundView : Model -> Html Msg
notFoundView model =
  div [] [
    h1 [] [ text "not found" ]
  ]

movieView : ApiResponse -> Html Msg
movieView model =
  filmView model

moviesView : List ApiResponse -> Html Msg
moviesView movies =
  div [
    classList [
      ("cf", True),
      ("pa2", True)
    ]
  ] (List.map filmView movies)


filmView : ApiResponse -> Html msg
filmView resp =
  let
      movie = resp.apiMovie
  in
    div [
        classList [
          ("fl", True),
          ("w-50", False),
          ("pa2", True),
          ("w-20", True),
          ("w-w-20-l", True)
        ]
      ]
      [ img [
          src movie.poster,
          classList [
            ("db", True),
            ("w-100", True),
            ("outline", True),
            ("black-10", True),
            ("dim", True)
          ]
        ] []
      , dl [
          classList [
            ("mt2", True),
            ("f6", True),
            ("lh-copy", True)
            ]
        ] [
          filmMetaData movie.title,
          filmMetaData movie.year,
          filmMetaData movie.kind
        ]
      ]

filmMetaData : String -> Html msg
filmMetaData data =
  dt [ class "m10 black truncate w-100" ] [ text data ]

-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
  Sub.none


-- HTTP
baseUrl : String
baseUrl = "http://localhost:8080/api/v0"

searchApi : Cmd Msg
searchApi =
  let
      url = baseUrl ++ "/movies"
  in
     Task.perform FetchFail MoviesFetchSucceed (Http.get decodeApiResponse url)

singleMovieSearch : String -> Cmd Msg
singleMovieSearch id =
  let
      url = String.join "/" [baseUrl, "movie", id]
  in
     Task.perform FetchFail MovieFetchSucceed (Http.get responseDecoder url)


decodeApiResponse : Json.Decoder (List ApiResponse)
decodeApiResponse =
  Json.list responseDecoder

type alias MovieResponse =
  { imdbID : String
  , poster : String
  , title : String
  , kind : String
  , year : String
  }

type alias ApiResponse =
  { title : String
  , fullPath : String
  , apiMovie : MovieResponse
  }

responseDecoder : Json.Decoder ApiResponse
responseDecoder =
  Json.object3 ApiResponse
    ("Title" := Json.string)
    ("FullPath" := Json.string)
    ("ApiMovie" := apiResponseDecoder)

apiResponseDecoder : Json.Decoder MovieResponse
apiResponseDecoder =
  Json.object5 MovieResponse
    ("ImdbID" := Json.string)
    ("Poster" := Json.string)
    ("Title" := Json.string)
    ("Type" := Json.string)
    ("Year" := Json.string)
