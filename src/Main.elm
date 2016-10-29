import Html exposing (..)
import Html.App as Html
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as Json exposing (..)
import Task
import Navigation
import Debug


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
      one = Debug.log "path" pathname
  in
     case pathname of
       "/src/index.html" -> Movies

       _ -> NotFound

urlParser : Navigation.Parser Route
urlParser =
  Navigation.makeParser parse

-- Model


type alias Model =
  { movies : List ApiResponse
  , route : Route
  }

init : Route -> (Model, Cmd Msg)
init route =
  (Model [] route , searchApi)

type Route
  = Movies
  | Movie String
  | NotFound

-- Update


type Msg
  = Search
  | FetchSucceed (List ApiResponse)
  | FetchFail Http.Error


update : Msg -> Model -> (Model, Cmd Msg)
update msg model =
  case msg of
    Search ->
      (model, searchApi)

    FetchSucceed movies ->
        ({ model | movies = movies }, Cmd.none)

    FetchFail _ ->
      (model, Cmd.none)


urlUpdate : Route -> Model -> (Model, Cmd Msg)
urlUpdate route model =
  ({model | route = route }, Cmd.none)


-- View


view : Model -> Html Msg
view model =
  case model.route of
    Movies ->
      moviesView model.movies

    _ ->
      notFoundView model

notFoundView : Model -> Html Msg
notFoundView model =
  div [] [
    h1 [] [ text "not found" ]
  ]

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

searchApi : Cmd Msg
searchApi =
  let
      url = "http://localhost:8080/api/v0/movies"
  in
     Task.perform FetchFail FetchSucceed (Http.get decodeApiResponse url)

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
